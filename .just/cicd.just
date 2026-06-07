#!/usr/bin/env just --justfile

[group("ci-cd")]
update-pinact:
    pinact run
