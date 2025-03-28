# Create new window in existing tmux session
tmux new-window -n "vdxDbg"
tmux split-window -v
tmux split-window -v
tmux send-keys -t :.0 './01_launch_dex.sh' C-m
tmux send-keys -t :.1 './02_go_run.sh' C-m
tmux send-keys -t :.2 './03_launch_npm_run_dev.sh' C-m
