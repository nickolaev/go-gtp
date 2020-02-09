#!/usr/bin/env bash

set -x

if [ "$EUID" -ne 0 ]
  then echo "Please run as root"
  exit
fi


EPC_COMPONENTS=(mme sgw pgw enb)

function cleanup {
    for c in ${EPC_COMPONENTS[@]}; do
        killall $c
    done
    tmux kill-session -t $SESSION
    exit
}
# Call the egress function
trap cleanup EXIT INT TERM

# Set Session Name
SESSION="GoEPC"
SESSIONEXISTS=$(tmux list-sessions | grep $SESSION)

# Only create tmux session if it doesn't already exist
if [ "$SESSIONEXISTS" = "" ]
then
    # Start New Session with our name
    tmux new-session -d -s $SESSION
    for i in ${!EPC_COMPONENTS[@]}; do
        c=${EPC_COMPONENTS[$i]}
        if [ $i != 0 ]; then
            tmux new-window -t $SESSION:$i
        fi 
        tmux rename-window -t $i "GoEPC ${c}"
        tmux send-keys -t $SESSION:$i "pushd $c; go run ." C-m
    done
fi

tmux attach-session -t $SESSION:0

# while ! tmux has-session -t $SESSION; do sleep 1; done
echo "Press ENTER to exit"
read
