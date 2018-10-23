import argparse
import requests
import re
import os
import subprocess
import json
import csv
import errno
from retry import retry
from subprocess import CalledProcessError

OUT_DIR = 'results'


class ReplicaCountMismatchException(Exception):
    """Raise when expected number of available replicas does not match the expected number of replicas"""
    pass

@retry(KeyError, tries=20, delay=5)
def get_ip_address(service='minibank'):
    command = ['kubectl', 'get', 'services', service, '-o', 'json']
    # allow exceptions to be raised, script should fail miserably in case of any errors
    command_out = subprocess.check_output(command)
    return json.loads(command_out)['status']['loadBalancer']['ingress'][0]['ip']


@retry(ReplicaCountMismatchException, tries=20, delay=5)
def wait_for_replicas(replicas, deployment='minibank'):
    command = ['kubectl', 'get', 'deployments', deployment, '-o', 'json']
    command_out = subprocess.check_output(command)
    ready_replicas = json.loads(command_out)['status']['readyReplicas']
    if ready_replicas != replicas:
        msg = "Expected Replicas: {}, Ready Replicas: {}".format(replicas, ready_replicas)
        print msg
        raise ReplicaCountMismatchException(msg)


def update_replica_count(replicas, deployment='minibank'):
    command = ['kubectl', 'scale', 'deployments', deployment, '--replicas', str(replicas)]
    command_out = subprocess.check_call(command)


def get_levels():
    levels = {
        10: 300,
        15: 300,
        20: 300,
        30: 300,
        50: 300,
        80: 300,
        100: 500,
        150: 500,
        200: 1000,
        250: 1000,
        300: 1000,
    }
    return levels


def main(endpoint, replica_counts, tag, payload):
    ip_address = get_ip_address()
    for replicas in replica_counts:
        update_replica_count(replicas)
        wait_for_replicas(replicas)
        print 'Collecting statistics for cluster with {} replicas'.format(replicas)
        results = []
        levels = get_levels()
        for con_level, req_count in levels.items():
            print 'Processing Concurrency = {}'.format(con_level)
            command_out = execute_ab(con_level, req_count, ip_address, payload, endpoint)
            level_results = {'con_level': con_level}
            for line in command_out.split('\n'):
                if len(level_results) == 1:
                    match = re.match('Requests per second:\s+([0-9\.]+)\s', line)
                    if match:
                        level_results['rps'] = float(match.group(1))
                elif len(level_results) == 2:
                    match = re.match('\s+98%\s+([0-9]+)', line)
                    if match:
                        level_results['98p'] = int(match.group(1))
                elif len(level_results) == 3:
                    match = re.match('\s+100%\s+([0-9]+)', line)
                    if match:
                        level_results['longest'] = int(match.group(1))
                else:
                    break
            results.append(level_results)

        with open('{}/{}_{}.csv'.format(OUT_DIR, tag, replicas), 'w') as csvfile:
            writer = csv.DictWriter(csvfile, fieldnames=['con_level', 'rps', '98p', 'longest'])
            writer.writeheader()
            for row in results:
                writer.writerow(row)

        # also dump the dictionary:
        with open('{}/{}_{}.json'.format(OUT_DIR, tag, replicas), 'w') as jsonfile:
            json.dump(results, jsonfile)


@retry(CalledProcessError, tries=5, delay=1)
def execute_ab(con_level, req_count, ip_address, payload, endpoint):
    """
    Since AB is notorious for not being reliable, wrap the subprocess call in a retry loop
    """
    command = [
        'ab',
        '-p',
        payload,
        '-T', 
        'application/json',
        '-m',
        'POST',
        '-n', str(req_count),
        '-c', str(con_level),
        '-s',
        '100',
        '-r',
        'http://{}:8080/{}'.format(ip_address, endpoint)]
    try:
        command_out = subprocess.check_output(command)
        return command_out
    except CalledProcessError:
        print "AB exited with error"
        raise

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Runs multiple AB tests for a Kubernetes endpoint and aggregates the results.')
    parser.add_argument('--endpoint', help='Endpoint')    
    parser.add_argument('--replicas', nargs='+', type=int, help='Replica counts')
    parser.add_argument('--tag', default='results', help='Tag to add to result filenames')
    parser.add_argument('--payload', help='File to use as payload')
    args = parser.parse_args()

    # verify payload exists:
    # a gnarly Traceback will be thrown if file does not exist or is not readable
    with open(args.payload, 'r') as file:
        pass

    if not os.path.exists(OUT_DIR):
        try:
            os.makedirs(OUT_DIR)
        except OSError as e:
            if e.errno != errno.EEXIST:
                raise

    main(args.endpoint, args.replicas, args.tag, args.payload)
