#!bin/sh
find . -type f -name "*.go" -exec sh -c 'echo "$(<License_txt.txt)

$(cat $0)" > $0' {} \;
echo "Text Added on start of all file"