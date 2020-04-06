curl -O https://artifacts.elastic.co/downloads/kibana/kibana-7.6.1-darwin-x86_64.tar.gz
curl https://artifacts.elastic.co/downloads/kibana/kibana-7.6.1-darwin-x86_64.tar.gz.sha512 | shasum -a 512 -c -
tar -xzf kibana-7.6.1-darwin-x86_64.tar.gz
cd kibana-7.6.1-darwin-x86_64/
