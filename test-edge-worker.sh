#!/bin/bash

for (( ; ; ))
do
    echo;
    echo curl -i https://akamai-test.lab.amplitude.com/ --connect-to ::akamai-test.lab.amplitude.com.edgekey-staging.net -H "$1" -H "Pragma: akamai-x-ew-debug-rp" -H "id: my-id"; echo;
    curl -i https://akamai-test.lab.amplitude.com/ --connect-to ::akamai-test.lab.amplitude.com.edgekey-staging.net -H "$1" -H "Pragma: akamai-x-ew-debug-rp" -H "id: my-id";
    echo;
done
