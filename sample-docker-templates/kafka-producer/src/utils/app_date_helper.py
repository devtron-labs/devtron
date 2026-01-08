import os
import pytz
from datetime import timedelta, datetime

def get_formatted_value(date_obj):
    return datetime.strftime(date_obj,"%Y-%m-%d %H:%M:%S")

def get_date_value(datetime):
    return datetime.strptime(datetime,"%Y-%m-%d %H:%M:%S").strftime("%Y-%m-%d")

def get_time_value(date_value):
    return datetime.strptime(date_value,"%Y-%m-%d %H:%M:%S").strftime("%H:%M:%S")

def get_time_valueHHMM(date_value):
    return datetime.strptime(date_value,"%Y-%m-%d %H:%M:%S").strftime("%H:%M")

def add_seconds(date_str, sec):
    date_obj = datetime.strptime(date_str,"%Y-%m-%d %H:%M:%S")
    time_change = timedelta(seconds=sec)
    return date_obj + time_change

def add_hours_to_dateobj(date_obj, hours):
    time_change = timedelta(hours=hours)
    return date_obj + time_change

def add_days_to_dateobj(date_obj, days):
    time_change = timedelta(days=days)
    return date_obj + time_change

def prev_date():
    return datetime.today() - timedelta(days=1)

def get_date_value(date_value):
    return datetime.strptime(date_value,"%Y-%m-%d %H:%M:%S").strftime("%Y-%m-%d")

def get_file_format(from_date, to_date):
    prefix = os.getenv("prefix_file_name")
    time_zone_name = os.getenv("time_zone_name")
    from_time = datetime.strptime(from_date,"%Y-%m-%d %H:%M:%S").strftime("%Y%m%d%H%M")
    to_time = datetime.strptime(to_date,"%Y-%m-%d %H:%M:%S").strftime("%Y%m%d%H%M")
    return f"{prefix}-{from_time}-{to_time}-{time_zone_name}"

def get_title_format(from_date, to_date, local_time_zone, time_zone):
    time_zone_name = os.getenv("time_zone_name")
    local_from_date = convert_timezone_date(from_date, local_time_zone)
    local_to_date = convert_timezone_date(to_date, local_time_zone)
    tz_from_date = convert_timezone_date(from_date, time_zone)
    tz_to_date = convert_timezone_date(to_date, time_zone)
    return f"{from_date} - {to_date} ({time_zone_name})          {tz_from_date} - {tz_to_date} (UTC)         {local_from_date} - {local_to_date} (IST)"

def convert_timezone_date(date_str, zone):
    time_zone = os.getenv("time_zone")
    date_obj = datetime.strptime(date_str,"%Y-%m-%d %H:%M:%S")
    timezone_obj = pytz.timezone(time_zone).localize(date_obj)
    conv_date = timezone_obj.astimezone(pytz.timezone(zone))
    conv_dt_str = conv_date.strftime("%Y-%m-%d %H:%M:%S")
    return conv_dt_str

def convert_from_to_timezone(date_str, from_zone, to_zone):
    date_obj = datetime.strptime(date_str,"%Y-%m-%d %H:%M:%S")
    timezone_obj = pytz.timezone(from_zone).localize(date_obj)
    conv_date = timezone_obj.astimezone(pytz.timezone(to_zone))
    conv_dt_str = conv_date.strftime("%Y-%m-%d %H:%M:%S")
    return conv_dt_str

def convert_timezone_from_UTC(date_str, zone):
    date_obj = datetime.strptime(date_str,"%Y-%m-%d %H:%M:%S")
    timezone_obj = pytz.timezone('UTC').localize(date_obj)
    conv_date = timezone_obj.astimezone(pytz.timezone(zone))
    conv_dt_str = conv_date.strftime("%Y-%m-%d %H:%M:%S")
    return conv_dt_str