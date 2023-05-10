go build 
scp -i "/home/manoelra/ec2-manoel.pem" data/* ec2-user@ec2-18-228-223-22.sa-east-1.compute.amazonaws.com:/home/ec2-user/data/
scp -i "/home/manoelra/ec2-manoel.pem" views/* ec2-user@ec2-18-228-223-22.sa-east-1.compute.amazonaws.com:/home/ec2-user/views/
scp -i "/home/manoelra/ec2-manoel.pem" luckytip ec2-user@ec2-18-228-223-22.sa-east-1.compute.amazonaws.com:/home/ec2-user/


