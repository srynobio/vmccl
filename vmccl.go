package main

import (
	"bufio"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/shenwei356/bio/seqio/fastx"
)

func main() {

	var args struct {
		Stdin  bool   `help:"Read from stdin."`
		Blob   string `help:"Blob text to hash using the SHA-512 algorithm."`
		VMC    bool   `help:"With output the result of the above blob/stdin base on the current VMC spec."`
		Fasta  string `help:"Will return VMC Sequence digest of this fasta file."`
		Length int    `help:"Length of digest id to return."`
	}
	args.Length = 24
	arg.MustParse(&args)

	if len(args.Fasta) > 1 {
		digestFasta(args.Fasta, args.Length)
	} else if len(args.Blob) > 1 {

		clean := scrubber(args.Blob)
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
		clean := scrubber(stdinText)
		toByte := []byte(clean)

		if args.VMC {
			fmt.Println(VMCDigest(toByte, args.Length))
		} else {
			fmt.Println(Digest(toByte, args.Length))
		}
	}
}

// --------------------------------------------------------------------- //

func scrubber(i string) string {
	cleanEnds := strings.TrimSpace(i)
	noSpace := strings.Replace(cleanEnds, " ", "", -1)

	return noSpace
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
