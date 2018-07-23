package main

import (
	"bufio"
	"compress/gzip"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/brentp/vcfgo"
	"github.com/brentp/xopen"
	"github.com/shenwei356/bio/seqio/fastx"
)

// Lookup map of chromosome -> VMC Seq_ID
var fastaVMC = make(map[string]string)

func main() {

	var args struct {
		Stdin   bool   `help:"Read from stdin."`
		Blob    string `help:"Blob text to hash using the SHA-512 algorithm."`
		Fasta   string `help:"Will return VMC Sequence digest ID of this fasta file."`
		VCF     string `help:"Will take input VCF file and updated to include VMC (sequence|location|allele) digest IDs."`
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

	// f, err := os.OpenFile("access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	//seqIDFile, err := os.OpenFile(fastaVMCFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if len(args.Fasta) > 1 && len(args.VCF) < 1 {
		// Open of append if fasta.vmc file exists.
		if _, err := os.Stat(fastaVMCFile); err != nil {
			seqIDFile, err := os.Create(fastaVMCFile)
			eCheck(err)
			defer seqIDFile.Close()
			digestFasta(args.Fasta, args.Length, seqIDFile)
		}
	} else if len(args.VCF) > 1 && len(args.Fasta) > 1 {

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
	} else if len(args.Blob) > 1 {

		clean := spaceScrubber(args.Blob)
		byteForm := []byte(clean)
		fmt.Println(Digest(byteForm, args.Length))

	} else if args.Stdin {
		scanner := bufio.NewScanner(os.Stdin)

		stdinText := ""
		for scanner.Scan() {
			stdinText += scanner.Text()
		}
		clean := spaceScrubber(stdinText)
		toByte := []byte(clean)
		fmt.Println(Digest(toByte, args.Length))

	} else {
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

		if seqID, ok := fastaVMC[variant.Chromosome]; ok {

			locationID := LocationDigest(seqID, variant)
			alleleID := AlleleDigest(locationID, variant)

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

// Non VMC URLEncoding hash.
func Digest(bv []byte, truncate int) string {
	hasher := sha512.New()
	hasher.Write(bv)

	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil)[:truncate])
	return sha
}

// --------------------------------------------------------------------- //

func LocationDigest(seqID string, vcfvar *vcfgo.Variant) string {

	intervalString := fmt.Sprintf("%d|%d", uint64(vcfvar.Start()-1), uint64(vcfvar.End()))
	location := fmt.Sprintf("<Location|%s|<Interval|%s>>", seqID, intervalString)
	DigestLocation := Digest([]byte(location), 24)

	locationID := fmt.Sprintf("VMC:GL_%s", DigestLocation)
	return locationID
}

// --------------------------------------------------------------------- //

func AlleleDigest(locationID string, vcf *vcfgo.Variant) string {

	state := strings.Join(vcf.Alt(), "")
	allele := fmt.Sprintf("<Allele|%s|%s>", locationID, state)
	DigestAllele := Digest([]byte(allele), 24)

	alleleID := fmt.Sprintf("VMC:GA_%s", DigestAllele)
	return alleleID
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

func spaceScrubber(i string) string {
	cleanEnds := strings.TrimSpace(i)
	noSpace := strings.Replace(cleanEnds, " ", "", -1)

	return noSpace
}

// --------------------------------------------------------------------- //

func eCheck(e error) {
	if e != nil {
		panic(e)
	}
	return
}

// --------------------------------------------------------------------- //
