#!/bin/bash

# First request (Valid CEP)
curl --request POST \
    --url http://localhost:8080/weather \
    --header 'Content-Type: application/json' \
    --data '{
	"cep": "70150900"
}'

# Second request (Invalid CEP)
curl --request POST \
    --url http://localhost:8080/weather \
    --header 'Content-Type: application/json' \
    --data '{
	"cep": "99999999"
}'

# Third request (Invalid CEP Format)
curl --request POST \
    --url http://localhost:8080/weather \
    --header 'Content-Type: application/json' \
    --data '{
	"cep": "123"
}'

# Third request (Invalid CEP Format)
curl --request POST \
    --url http://localhost:8080/weather \
    --header 'Content-Type: application/json' \
    --data '{
	"cep": "70150-900"
}'
