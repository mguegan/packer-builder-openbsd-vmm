# packer-builder-openbsd-vmm
[Packer](https://packer.io/) builder plugin for OpenBSD's VMM

## Talk
Find my BSDCan 2019 slides in https://github.com/double-p/presentations/tree/master/BSDCan/2019
Video not yet available.

## jumpstart
```
make install
packer build examples/openbsd.json
```
More details in BUILD.md

## bugs
Still some assumptions about how to the find the connected tap(4) interface

If you find something, please use ``make vmb'' and include the log.

# Remarks
This is heavily based on https://github.com/m110/packer-builder-hcloud and
https://github.com/prep/packer-builder-vmm
