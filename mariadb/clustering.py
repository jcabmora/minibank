#!/usr/bin/python3

"""
A simple bootstrap procedure for a mariadb/galera cluster
This only works for startup! It does not handle partitions,
and specialy, it does not handlle at all full cluster restarts
"""


import time
import socket
import netifaces
import subprocess
import logging
import os

from kazoo.client import KazooClient
from kazoo.retry import KazooRetry
from kazoo.exceptions import LockTimeout, NoNodeError

log = logging.getLogger(__name__)

LOCK = '/mariadb/bootstrap'
CLUSTER_NODES = '/mariadb/nodes'

def get_ip_address():
    """
    Return the first non-loopback IP address found by netifaces
    """
    for iface in netifaces.interfaces():
        if iface != 'lo':
            address_list = netifaces.ifaddresses(iface)
            if netifaces.AF_INET in address_list:
                try:
                    return address_list[netifaces.AF_INET][0]['addr']
                except Exception:
                    log.exception("Unable to determine ipaddress")


def get_zookeeper(hosts=None):
    """
    Waits and returns a connection to a Zookeeper ensemble
    """

    if not hosts:
        hosts = ['zookeeper']

    # wait forever or until a connection is successful
    while True:
        client = KazooClient(hosts=hosts, connection_retry=KazooRetry(max_tries=10, max_delay=5))
        try:
            client.start()
        except Exception:
            # we should probably handle different types of exceptions with different sleep times.
            time.sleep(3)
            continue
        return client


def main():
    """
    Starts the mysqld daemon
    Uses a zookeeper lock (global variable LOCK) to guarantee that only one node is started at a time.
    Uses a zookeeper node (global variable CLUSTER_NODES) to find other nodes that have joined the cluster
    Assumes that all the mysql nodes start their services using this same script
    """

    """
    IMPLEMENT YOUR LOGIC HERE
    """



def start_node(nodes):
    """
    Starts a secondary node as part of a mariadb galera cluster
    
    :param list nodes: a list of IP addresses of other nodes that have already joined the cluster
        For example, if the node has two nodes with IP addresses 172.16.0.1 and 172.16.1.2, then
        nodes will be ['172.16.0.1', '172.16.1.2']
    """
    server_id = len(nodes) + 1
    cmd = [
        'mysqld',
        '--user=mysql',
        '--wsrep-cluster-name=lab07',
        '--server-id={}'.format(server_id),
        '--wsrep_on=ON',
        '--binlog_format=ROW',
        '--wsrep_gtid_domain_id=1',
        '--wsrep_sst_method=rsync',
        '--wsrep_cluster_address=gcomm://{}'.format(','.join(nodes)),
        '--wsrep_provider=/usr/lib/galera/libgalera_smm.so'
    ]
    try:
        subprocess.Popen(cmd)
    except Exception:
        log.exception("Error while starting mysqld")


def bootstrap_cluster():
    """
    Bootstraps a mariadb galera cluster
    """
    cmd = [
        'mysqld',
        '--user=mysql',
        '--wsrep-cluster-name=lab07',
        '--server-id=1',
        '--wsrep_on=ON',
        '--wsrep-new-cluster',
        '--binlog_format=ROW',
        '--wsrep_gtid_domain_id=1',
        '--wsrep_sst_method=rsync',
        '--wsrep_cluster_address=gcomm://',
        '--wsrep_provider=/usr/lib/galera/libgalera_smm.so'
    ]
    try:
        subprocess.Popen(cmd)
    except Exception:
        log.exception("Error while bootstrapping mysqld")


def wait_until_primary():
    """
    Waits forever or until the node is in primary status
    """
    cmd = [
        'mysql',
        '-uroot',
        '-p{}'.format(os.environ['MYSQL_ROOT_PASSWORD']),
        '-e',
        'SHOW STATUS LIKE "wsrep_cluster_status"\G',
    ]
    while True:
        # if mysql is not running, a CalledProcessError will be raised.
        try:
            output = subprocess.check_output(cmd)
            status = output.decode().split('\n')[2].strip().replace('Value: ', '')
            if status == 'Primary':
                log.info("Node in WSREP status equals Primary")
                break
        except subprocess.CalledProcessError:
            log.exception("mysql command returned error")
        time.sleep(2)

if __name__ == '__main__':
    main()

