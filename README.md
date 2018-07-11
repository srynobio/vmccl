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

**Please keep in mind this tool only implements specfic aspects of the entire [VMC model](https://docs.google.com/document/d/12E8WbQlvfZWk5NrxwLytmympPby6vsv60RxCeD5wc1E/edit). A VMC Bundle and JSON file are not generated or validated.**

## Usage

```
Usage: vmccl [--stdin] [--blob BLOB] [--fasta FASTA] [--vcf VCF] [--logfile LOGFILE] [--length LENGTH]

Options:
  --stdin                Read from stdin.
  --blob BLOB            Blob text to hash using the SHA-512 algorithm.
  --fasta FASTA          Will return VMC Sequence digest ID of this fasta file.
  --vcf VCF              Will take input VCF file and updated to include VMC (sequence|location|allele) digest IDs.
  --logfile LOGFILE      Filename for output log file. [default: VMCCL.log]
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

## Description

Please review the [example](https://github.com/srynobio/vmccl/tree/vcf#examples) section for best practices instructions on how to run `vmccl`.

#### Fasta option:

`vmccl` will run the VMC digest algorithm on each record in the fasta file.  It will store the results into a file of the same name, with a `.vmc` extension added.  Subsequent runs of `vmccl` will check for the presence of the `fasta.vmc` file in the same location as the original fasta file.

The following is the format of the `fasta.vmc` file:

Leading Identifier (space separated) | VMC Seq ID | Description line of fasta |
-------------------------------------|------------|--------------------------|
1|VMC:GS\_jqi61wB\_nLCsUMtCXsS0Yau\_pKxuS21U|1 dna:chromosome chromosome:GRCh37:1:1:249250621:1

**The importance of using the correct fasta files to generate `VMC_GS` cannot be stressed enough as even a change of a single base will generate a completely different sequence identifier.  This is especially important when considering sharing `VMC_GA` with other institutions.** 

#### VCF option:

Please review the [example]() section for best practices instructions on how to run `vmccl`.

At this time to update a VCF file, an accompanying fasta file with a identical `Leading Identifier` is required.  If a `fasta.vmc` file has already been generated `vmccl` will look for it in the same location as the original fasta and collect the VMC_GS identifiers.

**Note:**

* Only VCFs which have ran [vt decompose](https://genome.sph.umich.edu/wiki/Vt#Decompose) will be accepted.
* If your VCF file contains sequence identifiers not found in the fasta file, the VCF record is printed to the new file without updated annotations.
* If your fasta file contains records not found in the VCF file they are skipped.
* Uses and implementation of the `fasta.vmc` file will change as the [seqrepo](https://github.com/biocommons/biocommons.seqrepo) becomes more widely available, and/when `vmccl` implements a SQL database backend.


An example of annotations added to the VCF file:

```
Added to the VCF header:
##INFO=<ID=VMCGAID,Number=1,Type=String,Description="VMC Allele identifier">
##INFO=<ID=VMCGLID,Number=1,Type=String,Description="VMC Location identifier">
##INFO=<ID=VMCGSID,Number=1,Type=String,Description="VMC Sequence identifier">

Added annotations to the VCF INFO field:
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

## Examples

The best practice method for adding VMC digest IDs to a VCF file are as follows:

1. First create a `fasta.vmc` file that will be used for all/future VCF updates.

    ```
    $> ./vmccl --fasta human_g1k_v37_decoy.fasta

    $> ls -l *fasta*
    human_g1k_v37_decoy.fasta 
    human_g1k_v37_decoy.fasta.vmc
    ```

    * In general creating the `fasta.vmc` file will take longer then adding  annotations to a VCF file, so pre-building it will decrease future VCF runtimes.
    * Please keep in mind the `.fasta` and `.fasta.vmc` file will need to be in the same location, or `vmccl` will rebuild the `.fasta.vmc` file.


2. Run `vmccl` on your vcf file.

```
$> ./vmccl --fasta human_g1k_v37_decoy.fasta --vcf clinvar_20180701.vcf.gz

$> ls -l *fasta* *vcf*
human_g1k_v37_decoy.fasta 
human_g1k_v37_decoy.fasta.vmc

clinvar_20180701.vcf.gz
clinvar_20180701.vmc.vcf.gz
```

* Output VCF will always be gzip compressed.

### TODO

Currency fasta generation utilizes parallel process.  Future releases will incorporate parallel process for VCF updating.


### BUGS AND LIMITATIONS

Please report any bugs or feature requests to the [issue tracker](https://github.com/srynobio/vmccl/issues)

AUTHOR Shawn Rynearson <shawn.rynearson@gmail.com>

LICENCE AND COPYRIGHT Copyright (c) 2018, Shawn Rynearson <shawn.rynearson@gmail.com> All rights reserved.

This module is free software; you can redistribute it and/or modify it under the same terms as GO itself.

DISCLAIMER OF WARRANTY BECAUSE THIS SOFTWARE IS LICENSED FREE OF CHARGE, THERE IS NO WARRANTY FOR THE SOFTWARE, TO THE EXTENT PERMITTED BY APPLICABLE LAW. EXCEPT WHEN OTHERWISE STATED IN WRITING THE COPYRIGHT HOLDERS AND/OR OTHER PARTIES PROVIDE THE SOFTWARE "AS IS" WITHOUT WARRANTY OF ANY KIND, EITHER EXPRESSED OR IMPLIED, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE. THE ENTIRE RISK AS TO THE QUALITY AND PERFORMANCE OF THE SOFTWARE IS WITH YOU. SHOULD THE SOFTWARE PROVE DEFECTIVE, YOU ASSUME THE COST OF ALL NECESSARY SERVICING, REPAIR, OR CORRECTION.

IN NO EVENT UNLESS REQUIRED BY APPLICABLE LAW OR AGREED TO IN WRITING WILL ANY COPYRIGHT HOLDER, OR ANY OTHER PARTY WHO MAY MODIFY AND/OR REDISTRIBUTE THE SOFTWARE AS PERMITTED BY THE ABOVE LICENCE, BE LIABLE TO YOU FOR DAMAGES, INCLUDING ANY GENERAL, SPECIAL, INCIDENTAL, OR CONSEQUENTIAL DAMAGES ARISING OUT OF THE USE OR INABILITY TO USE THE SOFTWARE (INCLUDING BUT NOT LIMITED TO LOSS OF DATA OR DATA BEING RENDERED INACCURATE OR LOSSES SUSTAINED BY YOU OR THIRD PARTIES OR A FAILURE OF THE SOFTWARE TO OPERATE WITH ANY OTHER SOFTWARE), EVEN IF SUCH HOLDER OR OTHER PARTY HAS BEEN ADVISED OF THE POSSIBILITY OF SUCH DAMAGES.)))
