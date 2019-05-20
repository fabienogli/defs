# DEFS

DEFS Encryption File System

# Supervisor/LoadBalancer Protocol
Using UDP 

Supervisor can send :
 - `0 size` for WhereTo : The LoadBalancer looks in every available storage and redirect to whichever it decides (with enough size) or send an error if not enough size available
 - `1 hash` for WhereIs : The LoadBalancer must lookup in the table and say "The file ided by `hash` is in `location` or send an error if not found"
 
 Examples :  
 Succesful WhereIs query    
 supervisor --> : `1 1def4334adsdhkj3`  
 loadbalancer --> : `0 10.10.10.1`
 
 Incorrect WhereIs query (file not found)  
 supervisor --> : `1 1def4334adsdhkj3`
 loadbalancer --> : '1' 
 
 Successful WhereTo query  
 supervisor --> : `0 120`
 loadbalancer --> : `0 10.10.10.1`
 
 Incorrect WhereTo query (not enough size)  
 supervisor --> : `0 120`
 loadbalancer --> : `2`  

|    Entity    | Code |                                 Args                                 |                                  Description                                 | Action/Response |
|:------------:|:----:|:--------------------------------------------------------------------:|:----------------------------------------------------------------------------:|:---------------:|
|  Supervisor  |   0  |  hash: The hash of the file    size : the size required in Kilobytes |           Asks the LB where to put a file of size `size` with hash           |      Action     |
| LoadBalancer |   0  |                                ip/dns                                |    The ip/dns of the storage where there is enough size to store the file    |     Response    |
| LoadBalancer |   1  |                                                                      | The hash is conflicting with an existing entry. A new hash must be generated |  Response/Error |
| LoadBalancer |   2  |                                                                      |            Not enough size available on any of the storage servers           |  Response/Error |
|  Supervisor  |   1  |   hash : The unique identifier of the file  (256 bit hash probably)  |                 Asks the LB where is the file of hash `hash`                 |      Action     |
| LoadBalancer |   0  |                                ip/dns                                |             The ip/dns of the storage server containing the file             |     Response    |
| LoadBalancer |   3  |                                                                      |                 The file was not found in any of the storage                 |  Response/Error |