#! /bin/bash

C=0
y=1
while getopts x:y: flag 
do
    case "${flag}" in
        x) x=${OPTARG}
           C=1
           ;;
        y) y=${OPTARG}
            ;;
    esac
done
if [[ "$C" == "1" ]] 
then
    rm Docker-compose.yaml;
    python3 dockerComposeCreator.py $x $y;
    docker-compose build;
    docker-compose up;
else
    echo "bash avvio.sh -x [numero_nodi] -y [numero_repliche]"
fi
