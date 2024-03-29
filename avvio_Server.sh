#! /bin/bash

C1=0
C2=1
y=1
z=8
while getopts x:y:z: flag 
do
    case "${flag}" in
        x) x=${OPTARG}
           C1=1
           ;;
        y) y=${OPTARG}
            ;;
        z) z=${OPTARG}
            C2=1
            ;;
    esac
done
if [[ "$C1" == "1" && "$C2" == "1" ]] 
then
    rm docker-compose.yml;
    python3 dockerComposeCreator.py $x $y $z;
    docker-compose build;
    docker-compose up;
else
    echo "bash avvio.sh -x [numero_nodi] [-y [numero_repliche]] [-z [Dimensione dell'anello]]"
fi
