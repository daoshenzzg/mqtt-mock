# build
cd ../mqtt-mock/src/com.mgtv.mqtt; go build mqtt-mock.go

# benchmark subscribe
./mqtt-mock -broker "tcp://127.0.0.1:8000" -c 200 -n 100000 -action sub

# benchmark publish
./mqtt-mock -broker "tcp://127.0.0.1:8000" -c 200 -n 100000 -size 1024 -action pub