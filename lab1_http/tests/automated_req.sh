directory="test_files"
port_number=8080
wait_time=1

echo -e "\e[34mStarting the POST request tests to the server!\e[0m"
echo
for file in "$directory"/*
do
    if [ "$file" != "$directory/doc_file.docx" ]; then
    echo "POST: $file"
    go run test.go $port_number POST $file
    sleep $wait_time
    echo
    else
        continue
    fi
    
done
echo -e "\e[32mPOST request test is done!\e[0m"

echo -e "\e[34mStarting the GET request test from the server!\e[0m"
echo
for file in "$directory"/*
do
    if [ "$file" != "$directory/doc_file.docx" ]; then
        fileName=$(basename "$file")    
        echo "GET: $fileName file from the server"
        go run test.go $port_number GET $fileName
        sleep $wait_time
        echo
    else
        continue
    fi
done
echo -e "\e[32mGET request test is done!\e[0m"

echo -e "\e[34mTesting Not Implemented HTTP request from client to the server like PUSH request!\e[0m"
echo
fileName=$directory/textfile.txt
go run test.go $port_number PUSH $fileName
sleep $wait_time
echo
echo -e "\e[32mThe Not Implemented request test is done!\e[0m"

echo -e "\e[34mTesting to POST Unsupported file extention to the server!\e[0m"
echo
fileName=doc_file.docx
go run test.go $port_number POST $directory/$fileName
sleep $wait_time
echo
echo -e "\e[32mThe posting of Unsupported file extention test is done!\e[0m"

echo -e "\e[31mAll the test cases is completed!!\e[0m"
