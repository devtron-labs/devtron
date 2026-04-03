import csv
import os
from src.utils.app_logger import init_logger, info, error

def write_csv_file(path, dict_data):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    keys = dict_data[0].keys()
    with open(path, 'w', newline='') as file:
        dict_writer = csv.DictWriter(file, keys, extrasaction='ignore', delimiter = ',')
        dict_writer.writeheader()
        dict_writer.writerows(dict_data)

def read_csv_file(path):
    header = []
    rows = []
    with open(path, 'r') as file:
        info("Reading file...")
        csvreader = csv.reader(file)
        headers = next(csvreader)
        for row in csvreader:
            if (len(row) > 0):
                new_dict={}
                count = 0
                for header in headers:
                    new_dict[header] = row[count]
                    count = count + 1
                rows.append(new_dict)
        info("End of reading file...")
    return rows
