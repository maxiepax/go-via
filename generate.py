import csv

# Sample data for the Go struct PoolForm
data = [
    {"name": "Pool1", "start_address": "192.168.1.10", "end_address": "192.168.1.50", "netmask": 24, "net_address": "192.168.1.0", "gateway": "192.168.1.1"},
    {"name": "Pool2", "start_address": "10.0.0.10", "end_address": "10.0.0.50", "netmask": 24, "net_address": "10.0.0.1", "gateway": "10.0.0.1"},
    {"name": "Pool3", "start_address": "172.16.0.10", "end_address": "172.16.0.50", "netmask": 16, "net_address": "10.0.0.1", "gateway": "172.16.0.1"},
    {"name": "Pool4", "start_address": "192.168.100.10", "end_address": "192.168.100.50", "netmask": 24, "net_address": "192.168.1.0", "gateway": "192.168.100.1"}
]

# Define CSV file path
file_path = 'pool_form_data.csv'

# Write data to CSV
with open(file_path, mode='w', newline='') as file:
    writer = csv.DictWriter(file, fieldnames=["name", "start_address", "end_address", "netmask", "net_address", "gateway"])
    writer.writeheader()
    writer.writerows(data)

dataHost = [
    {"fqdn": "machine1.domain.ch", "ip": "192.168.1.10", "mac": "80:3f:5d:00:0c:65"},
    {"fqdn": "hodor1.domain.ch", "ip": "192.168.1.13", "mac": "16:34:c3:14:11:90"},
    {"fqdn": "chasperli1.domain.ch", "ip": "192.168.1.14", "mac": "76:34:c3:13:11:90",},
    {"fqdn": "sennehund24.domain.ch", "ip": "192.168.1.15", "mac": "80:3f:5d:00:14:65"}
]

with open("hosts_form_data.csv", mode='w', newline='') as file:
    writer = csv.DictWriter(file, fieldnames=["fqdn", "ip", "mac"])
    writer.writeheader()
    writer.writerows(dataHost)