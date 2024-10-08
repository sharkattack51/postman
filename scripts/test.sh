echo "... test start "
dir=$(cd $(dirname $0) && pwd)
cd ${dir}/../cmd/postman
if [ "$1" = "" ]; then
    go test ./... -count=1
elif [ "$1" = "-v" ]; then
	go test -v ./... -count=1
elif [ "$1" = "-c" ]; then
	go test -cover ./... -coverprofile=${dir}/../test/cover.out
	go tool cover -html=${dir}/../test/cover.out -o ${dir}/../test/cover.html
fi
