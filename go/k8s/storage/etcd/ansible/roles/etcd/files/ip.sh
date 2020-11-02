# ip.sh hostname
ip=$(dig +short $1)
echo $ip
