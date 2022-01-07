# **KURE** - Kerbal User Repository

## Description

Kure is netkan.exe helper that allows easily create local repository with
`.netkan` packages, generate `.ckan` packages from them and pipe ready
packages to CKAN client.

Use cases:
- you are advanced ckan user and you want to modify existing netkan packages
or create your own.
- you are ckan metadata maintainer and you want to test your packages with
complex relationships without hassle.
- you want to package mod that can't be packaged
in official CKAN repo due to their
[de-indexing policy](https://github.com/KSP-CKAN/CKAN/blob/master/policy/de-indexing.md)

## Disclaimers

- this is quick and dirty code I wrote four years ago. I'm shocked it still works. Here be dragons...
- I don't care abou Windows so kure was tested only on Linux

## Workflow example

[![asciicast](https://asciinema.org/a/5eak9x445d0yfosmn1t8ixl2t.png)](https://asciinema.org/a/5eak9x445d0yfosmn1t8ixl2t)


## Install

`go install github.com/TeddyDD/kure@latest`
