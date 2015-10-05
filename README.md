Scanner FS
==========

Scanner FS is a FUSE filesystem that enumerates all of the hosts on a network device and displays them as readble files. Reading a file
will then trigger an nmap scan. In the future each host will instead be a directory, with a variety of scans available (nessus, nmap -o, etc...)
as readable files.

####USAGE:
    go get
    go build
    ./scannerfs en0 mnt
    ls mnt
    > 192.168.1.5
    cat mnt/192.168.1.5
    > nmap ...
