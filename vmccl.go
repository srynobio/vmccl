package main

import (
	"bufio"
	"compress/gzip"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/alexflint/go-arg"
	"github.com/brentp/vcfgo"
	"github.com/brentp/xopen"
	"github.com/shenwei356/bio/seqio/fastx"
)

// Lookup map of chromosome -> VMC Seq_ID
var fastaVMC = make(map[string]string)

func main() {

	var args struct {
		Fasta   string `help:"Will return VMC Sequence digest ID of this fasta file."`
		VCF     string `help:"Will take input VCF file and updated to include VMC (sequence|location|allele) digest IDs."`
		HGVS    string `help:"Valid HGVS expression to digest into VMC record." `
		LogFile string `help:"Filename for output log file."`
		Length  int    `help:"Length of digest id to return. MAX: 64"`
	}
	args.Length = 24
	args.LogFile = "VMCCL.log"
	arg.MustParse(&args)

	// VMC fasta record filename.
	fastaVMCFile := args.Fasta + ".vmc"

	// Creating log file.
	f, err := os.OpenFile(args.LogFile, os.O_RDWR|os.O_CREATE, 0755)
	eCheck(err)
	defer f.Close()
	log.SetOutput(f)

	switch {
	case len(args.Fasta) > 1 && len(args.VCF) > 1:
		// Open of append if fasta.vmc file exists.
		if _, err := os.Stat(fastaVMCFile); err != nil {
			seqIDFile, err := os.Create(fastaVMCFile)
			eCheck(err)
			defer seqIDFile.Close()
			digestFasta(args.Fasta, args.Length, seqIDFile)
		}
	case len(args.VCF) > 1 && len(args.Fasta) > 1:
		// check if .fasta.vmc exists
		if _, err := os.Stat(fastaVMCFile); err != nil {
			seqIDFile, err := os.Create(fastaVMCFile)
			eCheck(err)
			defer seqIDFile.Close()
			digestFasta(args.Fasta, args.Length, seqIDFile)
			digestVCF(args.VCF, args.Length)
		} else {
			seqIDFile, err := os.Open(fastaVMCFile)
			eCheck(err)
			defer seqIDFile.Close()

			updateFastaMap(seqIDFile)
			digestVCF(args.VCF, args.Length)
		}
	case len(args.HGVS) > 1 && len(args.Fasta) > 1:
		if _, err := os.Stat(fastaVMCFile); err != nil {
			seqIDFile, err := os.Create(fastaVMCFile)
			eCheck(err)
			defer seqIDFile.Close()
			digestFasta(args.Fasta, args.Length, seqIDFile)
			digestHGVS(args.HGVS)
		} else {
			seqIDFile, err := os.Open(fastaVMCFile)
			eCheck(err)
			defer seqIDFile.Close()

			updateFastaMap(seqIDFile)
			digestHGVS(args.HGVS)
		}
	default:
		panic("[ERROR] Required options not met.")
	}
	log.Println("vmccl finished!")
}

// --------------------------------------------------------------------- //

func digestFasta(file string, length int, wFile *os.File) {

	log.Printf("Creating digest for each record in %s", file)
	log.Printf("Fasta VMC file named: %s", wFile.Name())

	// Lifted from gofasta-vmc.go
	// Incoming fastq file.
	reader, err := fastx.NewDefaultReader(file)
	if err != nil {
		panic(err)
	}

	for chunk := range reader.ChunkChan(5000, 5) {
		if chunk.Err != nil {
			panic(chunk.Err)
		}

		for _, record := range chunk.Data {

			// create the VMC seq digest id and add VMC prefix.
			digestID := Digest(record.Seq.Seq, length)
			fastaSeqID := fmt.Sprintf("VMC:GS_%s", digestID)

			// get meta data for record.
			description := string(record.Name)
			splitDescription := strings.Split(description, " ")

			// update fasta map.
			fastaVMC[splitDescription[0]] = fastaSeqID

			writeRecord := fmt.Sprintf("%s|%s|%s\n", splitDescription[0], fastaSeqID, description)
			wFile.WriteString(writeRecord)
		}
	}
}

// --------------------------------------------------------------------- //

func digestVCF(file string, length int) {

	outFName := strings.Replace(file, "vcf", "vmc.vcf", -1)
	if strings.HasSuffix(outFName, "gz") == false {
		outFName = strings.Replace(outFName, "vcf", "vcf.gz", -1)
	}

	// open vcf file and read
	fh, err := xopen.Ropen(file)
	eCheck(err)
	defer fh.Close()

	rdr, err := vcfgo.NewReader(fh, false)
	eCheck(err)
	defer rdr.Close()

	// Add VMC INFO to the header.
	rdr.AddInfoToHeader("VMCGSID", "1", "String", "VMC Sequence identifier")
	rdr.AddInfoToHeader("VMCGLID", "1", "String", "VMC Location identifier")
	rdr.AddInfoToHeader("VMCGAID", "1", "String", "VMC Allele identifier")

	//create the new writer
	fi, err := os.OpenFile(outFName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	eCheck(err)
	defer fi.Close()

	gzipWriter := gzip.NewWriter(fi)
	defer gzipWriter.Close()

	writer, err := vcfgo.NewWriter(gzipWriter, rdr.Header)
	eCheck(err)

	log.Printf("Creating VMC records for VCF file: %s", file)
	log.Printf("Writing VCF records to file: %s", outFName)

	for {
		variant := rdr.Read()
		if variant == nil {
			break
		}

		// Check for alternate allele.
		altAllele := variant.Alt()
		if len(altAllele) > 1 {
			panic("Multi-allelic variant found, please pre-run vt decompose on VCF file.")
		}
		state := altAllele[len(altAllele)-1]

		if seqID, ok := fastaVMC[variant.Chromosome]; ok {

			locationID := LocationDigest(seqID, uint64(variant.Start()), uint64(variant.End()))
			alleleID := AlleleDigest(locationID, state)
			/////alleleID := AlleleDigest(locationID, variant.Alt())

			variant.Info().Set("VMCGSID", seqID)
			variant.Info().Set("VMCGLID", locationID)
			variant.Info().Set("VMCGAID", alleleID)
			writer.WriteVariant(variant)
		} else {
			log.Printf("Could not locate record for: %s in fasta file.", variant.Chromosome)
			writer.WriteVariant(variant)
		}
	}
}

// --------------------------------------------------------------------- //

func digestHGVS(hgvs string) {

	// Split the string into it's parts.
	hgvsInfo := strings.Split(hgvs, ":")

	var location []rune
	var seq []string
	for pos, x := range hgvsInfo[1] {

		// if first position and prefix is 'g'
		if pos == 0 && x != 103 {
			panic("Currently on genomic HGVS expression are digested.")
		}
		// dont need g or '.'
		if x == 103 || x == 46 {
			continue
		}
		if unicode.IsNumber(x) {
			location = append(location, x)
			continue
		}
		if unicode.IsLetter(x) {
			// Check for allowed
			switch {
			case x == 65:
			case x == 67:
			case x == 71:
			case x == 84:
			default:
				warn := fmt.Sprintf("Sequence %s out not allowed", string(x))
				panic(warn)
			}
			seq = append(seq, string(x))
			continue
		}
	}

	s := string(location)
	toInt, err := strconv.Atoi(s)
	eCheck(err)

	// Define needed elements
	start := uint64(toInt)
	end := uint64(toInt)
	state := seq[len(seq)-1]

	if seqID, ok := fastaVMC[hgvsInfo[0]]; ok {
		hgvsLocationDigest := LocationDigest(seqID, start, end)
		hgvsAlleleDigest := AlleleDigest(hgvsLocationDigest, state)
		fmt.Println(hgvsAlleleDigest)
	}
}

// --------------------------------------------------------------------- //

func LocationDigest(seqID string, start uint64, end uint64) string {

	intervalString := fmt.Sprintf("%d|%d", start-1, end)
	location := fmt.Sprintf("<Location|%s|<Interval|%s>>", seqID, intervalString)

	DigestLocation := Digest([]byte(location), 24)
	locationID := fmt.Sprintf("VMC:GL_%s", DigestLocation)
	return locationID
}

// --------------------------------------------------------------------- //

func AlleleDigest(locationID string, state string) string {

	allele := fmt.Sprintf("<Allele|%s|%s>", locationID, state)
	DigestAllele := Digest([]byte(allele), 24)

	alleleID := fmt.Sprintf("VMC:GA_%s", DigestAllele)
	return alleleID
}

// --------------------------------------------------------------------- //

// Non VMC URLEncoding hash.
func Digest(bv []byte, truncate int) string {
	hasher := sha512.New()
	hasher.Write(bv)

	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil)[:truncate])
	return sha
}

// --------------------------------------------------------------------- //

func updateFastaMap(file *os.File) {

	log.Printf("Found fasta VMC record file: %s", file.Name())

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fileText := scanner.Text()
		records := strings.SplitN(fileText, "|", 3)

		fastaVMC[records[0]] = records[1]
	}
}

// --------------------------------------------------------------------- //

func eCheck(e error) {
	if e != nil {
		panic(e)
	}
	return
}

// --------------------------------------------------------------------- //
