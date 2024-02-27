#!/usr/bin/env bash

tee -a /tmp/yamlls-input  | yaml-language-server --stdio | tee -a /tmp/yamlls-output 
