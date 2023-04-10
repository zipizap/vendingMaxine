#!/usr/bin/env bash
table="${1:-processing_engines}"
echo ">>>> TABLE: $table"
echo $table | awk '{printf ".header on\n.mode column\nSELECT * FROM %s;\n", $1}' | sqlite3 ./sqlite.db

