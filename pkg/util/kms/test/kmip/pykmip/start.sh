cd /work
if ! grep kmip.ciphertrustmanager.local /etc/hosts; then
   echo "0.0.0.0 kmip.ciphertrustmanager.local" >> /etc/hosts
fi
touch server.log
pykmip-server -f ./server.conf -l ./server.log &
tail -f server.log
