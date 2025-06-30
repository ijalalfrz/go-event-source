#!/bin/sh

export SIGNATURE_PRIVATE_KEY=$(cat ./test/api/config/signature_key_private.pem)
export SIGNATURE_PUBLIC_KEY=$(cat ./signature_key_public.pem)
bin/app http
