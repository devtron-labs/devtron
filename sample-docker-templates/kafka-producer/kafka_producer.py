import sys
import os
import json
import requests
from kafka import KafkaProducer
from kafka.errors import KafkaError
from datetime import datetime, timezone
from threading import Thread
import pandas as pd
from urllib.parse import urlparse

from src.utils.app_path import env_path, log_file_name
from dotenv import load_dotenv
load_dotenv()
load_dotenv(verbose=True)
load_dotenv(dotenv_path=env_path)

import src.config as conf
from src.utils.app_helper import time_diff, get_time_in_sec
from src.utils.app_logger import init_logger, info, error
from src.utils.app_csv import read_csv_file

# bootstrap_servers = json.loads(os.getenv("bootstrap_servers"))
topic = os.getenv("output_topic")

class MyThread(Thread):
  def __init__(self, instance, topic, rid, msg):
    Thread.__init__(self)
    self.instance = instance
    self.topic = topic
    self.rid = rid
    self.msg = msg

  def run(self):
    response = publish_message(self.instance, self.topic, self.rid, self.msg)
    if(response.is_done and response.succeeded()):
        info(f"{self.rid} published successfully")
    else:
        info(f"{self.rid} publish failed")

def get_producer():
    producer = None
    try:
        list_of_servers = os.getenv("bootstrap_servers")
        bootstrap_servers = list_of_servers.split(',')
        info(f"Bootstrap servers {bootstrap_servers}")
        producer = KafkaProducer(bootstrap_servers=bootstrap_servers,
                                value_serializer=lambda v: json.dumps(v).encode('utf-8'),
                                key_serializer=lambda k: json.dumps(k).encode('utf-8'))
    except Exception as ex:
        info('Exception while connecting Kafka', ex)
    finally:
        return producer                        

def publish_message(instance, topic, key, value):
    try:
        response = instance.send(topic, key=key, value=value)
        record_metadata = response.get(timeout=30)
        info(f"Key: {key} : Partition: {record_metadata.partition} Offset: {record_metadata.offset}")
        return response
    except KafkaError as kex:
        error(f"Kafka exception : {str(kex)}")
    except Exception as ex:
        error(f'Exception in publishing message {str(ex)}')

def utcformat(dt, timespec='milliseconds'):
    """convert datetime to string in UTC format (YYYY-mm-ddTHH:MM:SS.mmmZ)"""
    iso_str = dt.astimezone(timezone.utc).isoformat('T', timespec)
    return iso_str.replace('+00:00', 'Z')

def main():
    # args, parser = args_parse()
    src_path = os.getenv("src_path")
    log_file_path = os.path.join(src_path, "logs", log_file_name + '.log')
    init_logger(log_file=log_file_path, level=conf.logger_level)
    start_time = get_time_in_sec()
    try:
        info(f"Started at - {start_time}")
        src_type = os.getenv("src_type")
        file_size_limit = 3000000
        fsl = os.getenv("file_size_limit")
        if (fsl):
            file_size_limit = fsl
        dict_rows = []
        file_path = None
        if src_type == "file":
            file_name = os.getenv("file_name")
            file_path = os.path.join(src_path, "input", file_name)
        elif src_type == "url":
            url_path = os.getenv("url_path")
            a = urlparse(url_path)
            file_name = os.path.basename(a.path)
            # file_path = f"./tmp/{file_name}"
            file_path = os.path.join(src_path, "urltmp", file_name)
            df = pd.read_csv(url_path)
            df.head()
            df.to_csv(file_path,index=False)
        else:
            info(f"Unsupported source type ({src_type}) value")

        if (file_path):
            info(f"Processing file {file_path}")
            file_size = os.path.getsize(file_path)
            info(f"File size: {file_size}, File size limit: {file_size_limit}")
            if (file_size <= int(file_size_limit)):
                dict_rows = read_csv_file(file_path)
            else:
                info(f"{file_path} file size exceeds {file_size} limit")
            if (len(dict_rows) > 0):
                info(f"Number of RIDs to process: {len(dict_rows)}")
                process_rids(dict_rows)
            else:
                info(f"No records found in the source file to process")
    except Exception as ex:
        error(f'Exception while producing message into Kafka : {str(ex)}')
    finally:
        prev_time, prstr = time_diff(start_time)
        info(f"Ended at - {get_time_in_sec()}")
        info("Over all total time taken: " + prstr)
        sys.exit(0)

def process_rids(dict_rows):
    instance = get_producer()
    info(f"Kafka instance: {instance}")
    if (instance):
        count = 0
        thread_count = 10
        tc = os.getenv("thread_count")
        if (tc):
            thread_count = int(tc)
        thread_list = []
        for dict_row in dict_rows:
            count = count + 1
            rid = dict_row["RID"]
            reg_type = dict_row["REG_TYPE"]
            info(f"Start processing {rid}")
            msg = get_message(rid, reg_type, topic)
            info(f"Message: {msg}")
            thread_list.append(MyThread(instance, topic, rid, msg))
            if (count == thread_count):
                process_thread(thread_list)
                count = 0
                thread_list = []

        if (len(thread_list) > 0):
            process_thread(thread_list)

def process_thread(thread_list):
    for thread in thread_list:
        thread.start()

    for thread in thread_list:
        thread.join()

def get_message(rid, reg_type, topic):
    now = datetime.now(tz=timezone.utc)
    utc_now = utcformat(now)
    message_bus_address = os.getenv("message_bus_address")
    if not message_bus_address:
        message_bus_address = topic
    return {
        "reg_type":reg_type,
        "rid":rid,
        "isValid":False,
        "internalError":False,
        "messageBusAddress":{
            "address":message_bus_address
        },
        "retryCount":None,
        "tags":{},
        "lastHopTimestamp":utc_now}

if __name__ == "__main__":
    main()


