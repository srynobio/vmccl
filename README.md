<div align="center">
<img src="https://github.com/srynobio/vmccl/blob/vcf/vmccllogo.png"><br><br>
</div>

## Variation Model Collaboration Command Line tool.
#### Version: 1.1.2

Lightweight command line implementation of the VMC SHA-512 algorithm to allow rapid digest generation.

The follow input types are currently allowed:

* Fasta
* VCF
* STDIN
* text blobs.

**Please keep in mind this tool only implements the VMC digest aspects of the entire [VMC model](https://docs.google.com/document/d/12E8WbQlvfZWk5NrxwLytmympPby6vsv60RxCeD5wc1E/edit). A VMC Bundle and JSON file are not generated or validated.**

## Usage

```
$> vmccl --help
Usage: vmccl [--stdin] [--blob BLOB] [--fasta FASTA] [--vcf VCF] [--logfile LOGFILE] [--length LENGTH]

Options:
  --stdin                Read from stdin.
  --blob BLOB            Blob text to hash using the SHA-512 algorithm.
  --fasta FASTA          Will return VMC Sequence digest of this fasta file.
  --vcf VCF              Will take input VCF file and updated to include VMC digest IDs. Option Requires fasta or fasta.vmcseq file.
  --logfile LOGFILE      Filename for output log file. [default: vmccl.log]
  --length LENGTH        Length of digest id to return. MAX: 64 [default: 24]
  --help, -h             display this help and exit
```

## Installing

Easiest method to run `vmccl` is to download the most recent executable corresponding to your computer environment.

*  Download

OS | Platform | Release
---|---|---
darwin | amd64 | [darwin](https://github.com/srynobio/vmccl/releases)
linux | amd64 | [linux](https://github.com/srynobio/vmccl/releases)

* Then run (example)

```
 $> wget <release link> .
 $> mv vmccl_linux64 vmccl
 $> chmod a+x vmccl

```

**Additional builds and features can be requested [here](https://github.com/srynobio/vmccl/issues)**

## Examples:

#### Fasta option:

`vmccl` will run the VMC digest algorithm on each record in the fasta file.  It will store the results into a file of the same name, with a `.vmc` extension added.  `vmccl` will check for the presence of this file in the same location as the original for future operations.

```
$> vmccl --fasta Chr1-GRCh37.fasta
$> cat Chr1-GRCh37.fasta.vmc

1|VMC:GS_jqi61wB_nLCsUMtCXsS0Yau_pKxuS21U|1 dna:chromosome chromosome:GRCh37:1:1:249250621:1
```

#### VCF option:

At this time, to update a VCF file an accompanying fasta file with identical unique sequence identifiers is required.  If a `fasta.vmc` file has already been generated `vmccl` will look for it in the same location as the original fasta file and use it for VMC_GS identifiers.

**Note:**

* If your VCF file contains sequence identifiers not found in the fasta file, the record is printed to the new file without updated annotations.
* If your fasta file contains records not found in the VCF file they are skipped.
* Uses and implementation of the `fasta.vmc` record file may change in the future as the [seqrepo](https://github.com/biocommons/biocommons.seqrepo) becomes more widely used.


`vmccl` will parse a valid VCF file and add the following annotations:

```
Added to the VCF header:
##INFO=<ID=VMCGAID,Number=1,Type=String,Description="VMC Allele identifier">
##INFO=<ID=VMCGLID,Number=1,Type=String,Description="VMC Location identifier">
##INFO=<ID=VMCGSID,Number=1,Type=String,Description="VMC Sequence identifier">

Added annotations to the VCF record:
1       949523  183381  C       T       .       .       ALLELEID=181485;CLNDISDB=MedGen:C4015293,OMIM:616126,Orphanet:ORPHA319563;CLNDN=Immunodeficiency_38_with_basal_ganglia_calcification;CLNHGVS=NC_000001.10:g.949523C>T;CLNREVSTAT=no_assertion_criteria_provided;CLNSIG=Pathogenic;CLNVC=single_nucleotide_variant;CLNVCSO=SO:0001483;CLNVI=OMIM_Allelic_Variant:147571.0003;GENEINFO=ISG15:9636;MC=SO:0001587|nonsense;ORIGIN=1;RS=786201005;VMCGSID=VMC:GS_jqi61wB_nLCsUMtCXsS0Yau_pKxuS21U;VMCGLID=VMC:GL_VMC:GS_UqMzt_PvRNhrFl31m8N7SbCGdDpmAtsp;VMCGAID=VMC:GA_VMC:GS_-sajfzQq1Q_PfOAPMPQRodzFclkX8ksp
```

#### STDIN option:

All `stdin` and `blob` text will have newlines and spaces removed.

```
$> wc -lmw irobot.txt
6041   70216  403176 irobot.txt

$> cat irobot.txt | vmccl --stdin
VMC:GS_mbeo1K0MZwIHizAurCs2hYwA7LMyXSX0

$> cat irobot.txt | vmccl --stdin --length 60
VMC:GS_mbeo1K0MZwIHizAurCs2hYwA7LMyXSX0OQtVedGfpGBChEOV4jv58F1SeXpq0K5rUGsytqHm4_1oicIh

```

#### Blob option:

```
$> vmccl --blob "I, Robot Isaac Asimov. TO JOHN W. CAMPBELL, JR, who godfathered THE ROBOTS"
VMC:GS_p6WvpVcb0_hJj5Y_4Za3o01Ln40R-Ijz

```
