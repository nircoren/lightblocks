<b>Quick Start</b>: <br />
	1. run in cmd: $ git clone https://github.com/nircoren/lightblocks.git <br />
	2. put .env i sent you on root dir. make sure its called .env <br />
	3. build images: $ docker-compose build
	4. run server: $ docker-compose up -d server
	5. run client (remove container after finish): $ docker run --rm -it lightblocks-client
	6. first prompt will be username, second should be a json in this format:
		[
			{
				"Action": "addItem",
				"Key": "1",
				"Value": "val1"
			},
			{
				"Action": "getAllItems"
			},
			{
				"Action": "addItem",
				"Key": "2",
				"Value": "val2"
			}
		]
<br/><br/>
  !!! might have issue with parsing the json on bash/wt in windows, better to use another terminal. <br /> <br />
  
    Example multiple inputs:


	You can try to run multiple instances of the program with different inputs:



<b> My assumptions during project: </b> <br />
	Don't need to remove command from map execute making it. <br />
	Client shouldn't be a server. <br />
	Order of actions should stay the same for each client, but doesnt matter if order is not the same between 2 different clients <br />
	Shouldn't cancel sending and receiveing based on one bad message. <br />

