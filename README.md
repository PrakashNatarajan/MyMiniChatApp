# MyMiniChatApp

#Build ReceiverServer Image

docker build --network host -t receserver:0.1 ReceServer/


#Build SenderServer Image

docker build --network host -t sendserver:0.1 SendServer/


#Run ReceiverServer Image

docker run --rm --name receserver -p 5600:5600 -d receserver:0.1


#Run SenderServer Image

docker run --rm --name sendserver -p 6500:6500 -d sendserver:0.1

