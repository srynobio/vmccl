## Variation Model Collaboration Command Line tool.

Lightweight command line implementation of the VMC SHA-512 algorithm to allow rapid digest generation and comparison from fasta files, STDIN or text blob.

**Please keep in mind this tool only implements the VMC Sequence digest aspect of the entire [VMC model](https://docs.google.com/document/d/12E8WbQlvfZWk5NrxwLytmympPby6vsv60RxCeD5wc1E/edit). A VMC Bundle and JSON file is not generated or validated.**

```
$> vmccl -h

Usage: vmccl [--stdin] [--blob BLOB] [--vmc] [--fasta FASTA] [--length LENGTH]

Options:
  --stdin                Read from stdin.
  --blob BLOB            Blob text to hash using the SHA-512 algorithm.
  --vmc                  With output the result of the above blob/stdin base on the current VMC spec.
  --fasta FASTA          Will return VMC Sequence digest of this fasta file.
  --length LENGTH        Length of digest id to return. [default: 24]
  --help, -h             display this help and exit
```

Adding the `--vmc` option will return a base64 URL Encoded ID. All other options
are self-explanatory with examples shown below.

## Installing

Easiest method to run `vmccl` is to download the executable corresponding to your computer environment.

*  Download

OS | Platform | Link
---|---|---
darwin | amd64 | [darwin](https://github.com/srynobio/vmccl/releases)
linux | amd64 | [linux](https://github.com/srynobio/vmccl/releases)

* Then run (example)

```
 $> wget https://github.com/srynobio/vmccl/releases/download/v1.0.0/vmccl_linux64 .
 $> mv vmccl_linux64 vmccl
 $> chmod a+x vmccl

```


## Examples:

All `--stdin` and `blob` text will have newlines and spaces removed.

#### stdin option:

This example of stdin show the output with and without the `--vmc` option added

```
$> wc -lmw irobot.txt
6041   70216  403176 irobot.txt

$> cat irobot.txt | vmccl --stdin
mbeo1K0MZwIHizAurCs2hYwA7LMyXSX0

$> cat irobot.txt | vmccl --stdin --vmc
mbeo1K0MZwIHizAurCs2hYwA7LMyXSX0

$> cat irobot.txt | vmccl --stdin --length 60
mbeo1K0MZwIHizAurCs2hYwA7LMyXSX0OQtVedGfpGBChEOV4jv58F1SeXpq0K5rUGsytqHm4/1oicIh

$> cat irobot.txt | vmccl --stdin --vmc --length 60
mbeo1K0MZwIHizAurCs2hYwA7LMyXSX0OQtVedGfpGBChEOV4jv58F1SeXpq0K5rUGsytqHm4_1oicIh

```

#### blob option:

```
$> vmccl --blob "I, Robot Isaac Asimov. TO JOHN W. CAMPBELL, JR, who godfathered THE ROBOTS"
p6WvpVcb0/hJj5Y/4Za3o01Ln40R+Ijz

$> vmccl --blob "I, Robot Isaac Asimov. TO JOHN W. CAMPBELL, JR, who godfathered THE ROBOTS" --vmc
p6WvpVcb0_hJj5Y_4Za3o01Ln40R-Ijz

```

#### fasta option:

```
$> vmccl.go --fasta NC_000019.10.fasta

Description line:  NC_000019.10 Homo sapiens chromosome 19, GRCh38.p7 Primary Assembly
VMCDigest ID:  IIB53T8CNeJJdUqzn9V_JnRtQadwWCbl
```
