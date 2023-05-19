for i in $(./busybox --list); do 
  ln -sv busybox $i 
done
