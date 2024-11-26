# http_server
- The server is run by typing "./http_server.exe port1"
- The server requires there to be a folder called resources next to the .exe file that contains all the files that is posted/fetched.
# proxy
- The proxy is run by typing "./proxy.exe port2"
- An example command you can use is "curl.exe -X GET localhost:port1/file_name -x localhost:port2" after running the http_server.
