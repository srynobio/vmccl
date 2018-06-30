package main

import (
	"bufio"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
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
		Stdin  bool   `help:"Read from stdin."`
		Blob   string `help:"Blob text to hash using the SHA-512 algorithm."`
		VMC    bool   `help:"With output the result of the above blob/stdin base on the current VMC spec."`
		Fasta  string `help:"Will return VMC Sequence digest of this fasta file."`
		VCF    string `help:"Will take input VCF file and updated to include VMC digest IDs. Option Requires fasta or fasta.vmcseq file."`
		Length int    `help:"Length of digest id to return."`
	}
	args.Length = 24
	arg.MustParse(&args)

	if len(args.Fasta) > 1 && len(args.VCF) < 1 {
		digestFasta(args.Fasta, args.Length)
	} else if len(args.VCF) > 1 && len(args.Fasta) > 1 {

		fastaVMCFile := args.Fasta + ".vmc"
		// check if .fasta.vmc exists
		if _, err := os.Stat(fastaVMCFile); err != nil {
			seqIDFile, err := os.Create(fastaVMCFile)
			eCheck(err)
			defer seqIDFile.Close()
			digestFastaVCF(args.Fasta, args.Length, seqIDFile)
		}
		digestVCF(args.VCF, args.Length)
	} else if len(args.Blob) > 1 {

		clean := spaceScrubber(args.Blob)
		byteForm := []byte(clean)

		if args.VMC {
			fmt.Println(VMCDigest(byteForm, args.Length))
		} else {
			fmt.Println(Digest(byteForm, args.Length))
		}

	} else if args.Stdin {
		scanner := bufio.NewScanner(os.Stdin)

		stdinText := ""
		for scanner.Scan() {
			stdinText += scanner.Text()
		}
		clean := spaceScrubber(stdinText)
		toByte := []byte(clean)

		if args.VMC {
			fmt.Println(VMCDigest(toByte, args.Length))
		} else {
			fmt.Println(Digest(toByte, args.Length))
		}
	} else {
		panic("[ERROR] Required options not met.")
	}

}

// --------------------------------------------------------------------- //

func spaceScrubber(i string) string {
	cleanEnds := strings.TrimSpace(i)
	noSpace := strings.Replace(cleanEnds, " ", "", -1)

	return noSpace
}

// --------------------------------------------------------------------- //

func digestFastaVCF(inFile string, length int, wFile *os.File) {

	// Lifted from gofasta-vmc.go
	// Incoming fastq file.
	reader, err := fastx.NewDefaultReader(inFile)
	if err != nil {
		panic(err)
	}

	for chunk := range reader.ChunkChan(5000, 5) {
		if chunk.Err != nil {
			panic(chunk.Err)
		}

		for _, record := range chunk.Data {

			digestID := VMCDigest(record.Seq.Seq, length)
			description := string(record.Name)
			splitDescription := strings.Split(description, " ")

			// update fasta map.
			fastaVMC[splitDescription[0]] = digestID

			writeRecord := fmt.Sprintf("%s|%s|%s\n", splitDescription[0], digestID, description)
			wFile.WriteString(writeRecord)
		}
	}
}

// --------------------------------------------------------------------- //

func digestFasta(file string, length int) {

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
			digestID := VMCDigest(record.Seq.Seq, length)

			fmt.Println("Description line: ", string(record.Name))
			fmt.Println("VMCDigest ID: ", digestID)
		}
	}
}

// --------------------------------------------------------------------- //
/*
for key, value := range fastaVMC {
	fmt.Println("Key:", key, "Value:", value)
}
*/

// --------------------------------------------------------------------- //

func digestVCF(file string, length int) {

	outFile := strings.Replace(file, "vcf", "vmc.vcf", -1)

	fh, err := xopen.Ropen(file)
	eCheck(err)
	defer fh.Close()

	// create the writer
	output, err := os.Create(outFile)
	eCheck(err)
	defer output.Close()

	// VCF reader
	rdr, err := vcfgo.NewReader(fh, false)
	eCheck(err)
	defer rdr.Close()

	// Add VMC INFO to the header.
	rdr.AddInfoToHeader("VMCGSID", "1", "String", "VMC Sequence identifier")
	rdr.AddInfoToHeader("VMCGLID", "1", "String", "VMC Location identifier")
	rdr.AddInfoToHeader("VMCGAID", "1", "String", "VMC Allele identifier")

	//create the new writer
	writer, err := vcfgo.NewWriter(output, rdr.Header)
	eCheck(err)

	for {
		variant := rdr.Read()
		if variant == nil {
			break
		}

		// Check for alternate allele.
		altAllele := variant.Alt()
		if len(altAllele) > 1 {
			panic("multiallelic variant found, please pre-run with vt.")
		}

		seqID := fastaVMC[variant.Chromosome]
		locationID := LocationDigest(seqID, variant)
		alleleID := AlleleDigest(locationID, variant)

		variant.Info().Set("VMCGSID", seqID)
		variant.Info().Set("VMCGLID", locationID)
		variant.Info().Set("VMCGAID", alleleID)
		writer.WriteVariant(variant)
	}
}

// --------------------------------------------------------------------- //

// Non VMC URLEncoding hash.
func Digest(bv []byte, truncate int) string {
	hasher := sha512.New()
	hasher.Write(bv)

	sha := base64.StdEncoding.EncodeToString(hasher.Sum(nil)[:truncate])
	return sha
}

// --------------------------------------------------------------------- //

// VMC implemented digest
func VMCDigest(bv []byte, truncate int) string {
	hasher := sha512.New()
	hasher.Write(bv)

	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil)[:truncate])
	vmcsha := "VMC:GS_" + sha
	return vmcsha
}

// --------------------------------------------------------------------- //

func LocationDigest(seqID string, vcfvar *vcfgo.Variant) string {

	intervalString := fmt.Sprintf("%d:%d", uint64(vcfvar.Start()-1), uint64(vcfvar.End()))
	location := fmt.Sprintf("<Location|%s|<Interval|%s>>", seqID, intervalString)
	DigestLocation := Digest([]byte(location), 24)

	locationID := fmt.Sprintf("VMC:GL_%s", DigestLocation)
	return locationID
}

// --------------------------------------------------------------------- //

func AlleleDigest(locationID string, vcf *vcfgo.Variant) string {

	state := fmt.Sprint(vcf.Alt())

	allele := fmt.Sprintf("<Allele:<Identifier:%s>:%s>", locationID, state)
	//	allele := "<Allele:<Identifier:" + v.Location.id + ">:" + state + ">"
	DigestAllele := Digest([]byte(allele), 24)
	alleleID := fmt.Sprintf("VMC:GA_%s", DigestAllele)
	return alleleID
}

// --------------------------------------------------------------------- //

func eCheck(e error) {
	if e != nil {
		panic(e)
	}
	return
}

// --------------------------------------------------------------------- //
