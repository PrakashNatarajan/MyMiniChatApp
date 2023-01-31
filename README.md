# MyMiniChatApp

#Build ReceServer Image

docker build --network host -t receserver:0.1 ReceServer/


#Build SendServer Image

docker build --network host -t sendserver:0.1 SendServer/


#Run ReceServer Image

docker run --rm --name receserver -p 5600:5600 -d receserver:0.1


#Run SendServer Image

docker run --rm --name sendserver -p 6500:6500 -d sendserver:0.1

