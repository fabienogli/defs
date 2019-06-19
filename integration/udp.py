import socket
import sys


def send(message):
    # Sending data
    print('sending {!r}'.format(message))
    sent = sock.sendto(message, server_adress)


def receive():
    # Receive response
    print('waiting to receive')
    data, server = sock.recvfrom(4096)
    return format(data)

    
def main():
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    server_adress = ('localhost', 8080)
    message = b'This is the message. It will be repeated.'
    try:
        send(message)
        received = receive()
        print('received {!r}'.format(data))

    except:
        print "Unexpected error:", sys.exc_info()[0]
    finally:
        print('closing socket')
        sock.close()

