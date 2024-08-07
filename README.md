!! made change to folder structure and relations between pkg and send, receive on 07/08/ 13:00 (after sending solution to hr) !!
<b>
</b></b>

Quick Start</b>: <br />
   1. $ git clone https://github.com/nircoren/lightblocks.git <br />
   2. put .env i sent you on root dir. make sure its called .env <br />
   3. $ docker-compose up --build <br />
  4. $ docker exec <containter_name> /app/bin/client --username <username> --msgs <msgs> <br />
	docker exec sends the messages to client
<br/><br/>
        <b>example input:</b> <br/>
	$ docker exec lightblocks-client-1 /app/bin/client --username nir --msgs '[{"command":"addItem","key":"key1","value":"value1"},{"command":"addItem","key":"key2","value":"value2"},{"command":"addItem","key":"key3","value":"value3"},{"command": "addItem","key":"key111","value":"yaythere"},{"command":"getAllItems"}]'  <br /> <br />
  !!! might have issue with parsing the json on bash/wt in windows, better to use another terminal. <br /> <br />
	
	 Testing:  <br />
	 $ docker exec -it lightblocks-server-1 /bin/bash <br />
	 $ go test ./...
	 <br />
  
	    When making cli request, docker exec,  you need to pass --username and --message in this format: <br />
	    username string
	    type Message struct {
	      Command string `json:"command"`
	      Key     string `json:"key,omitempty"`
	      Value   string `json:"value,omitempty"`
	    }

    example multiple users: <br /> https://github.com/user-attachments/files/16510323/example_multiple_users.txt


<b> My assumptions during project: </b> <br />
	Don't need to remove command from map execute making it. <br />
	Client shouldn't be a server. <br />
	Order of actions should stay the same for each client, but doesnt matter if order is not the same between 2 different clients <br />
	Shouldn't cancel sending and receiveing based on one bad message. <br />

