import datetime as dt
import hashlib
import json
import pprint
import re
import time
import traceback
import src.utils.app_logger as app_logger

def sha256_hash(data):
    '''
    data assumed as type bytes
    '''
    m = hashlib.sha256()
    m.update(data)
    h = m.hexdigest().upper()
    return h

def read_token(response):
    cookies = response.headers['Set-Cookie'].split(';')
    for cookie in cookies:
        key = cookie.split('=')[0]
        value = cookie.split('=')[1]
        if key == 'Authorization':
            return value
    return None

def get_timestamp(seconds_offset=None):
    '''
    Current TS.
    Format: 2019-02-14T12:40:59.768Z  (UTC)
    '''
    delta = dt.timedelta(days=0)
    if seconds_offset is not None:
        delta = dt.timedelta(seconds=seconds_offset)

    ts = dt.datetime.utcnow() + delta
    ms = ts.strftime('%f')[0:3]
    s = ts.strftime('%Y-%m-%dT%H:%M:%S') + '.%sZ' % ms
    print(s)
    return s

def is_str(s):
    return isinstance(s, str)

def printResponse(r, h=None):
    print('Status code =  %s' % r.status_code)
    if h is not None:
        print('Headers =  %s' % r.headers)
    print('Links =  %s' % r.links)
    print('Encoding = %s' % r.encoding)
    print('Response Data = %s' % r.content)
    print('Size = %s' % len(r.content))

def responseToDict(r):
    try:
        app_logger.myprint('Response: <%d>' % r.status_code)
        r = r.content.decode()  # to str
        r = json.loads(r)
    except:
        r = traceback.format_exc()

    return r

def keyExists(key, d):
    if key in d.keys():
        return True
    else:
        return False

def Pprint(s):
    return s if isinstance(s, str) else pprint.pformat(s)


def Wait(t):
    app_logger.myprint("Waiting for " + Pprint(t) + " seconds")
    time.sleep(t)


def match(reg, st):
    regex = r".*(%s).*" % re.escape(reg)
    m = re.match(regex, st, re.DOTALL)
    return True if m else False


def rid_to_center_timestamp(rid):
    center_id = rid[:5]
    timestamp = rid[-14:-10] + '-' + rid[-10:-8] + '-' + rid[-8:-6] + 'T' + rid[-6:-4] + ':' + rid[-4:-2] + ':' + rid[
                                                                                                                  -2:] + ".000Z"
    return center_id, timestamp


def get_time_in_sec():
    return round(time.time() * 1000)


def time_diff(prev_millis):
    curr_millis = get_time_in_sec()
    diff = curr_millis - prev_millis
    seconds, milliseconds = divmod(diff, 1000)
    minutes, seconds = divmod(seconds, 60)
    hours, minutes = divmod(minutes, 60)
    days, hours = divmod(hours, 24)
    return curr_millis, pPrint(hours) + " hours, " + pPrint(minutes) + " minutes, " + pPrint(
        seconds + milliseconds / 1000) + " seconds"


def pPrint(s):
    return s if isinstance(s, str) else pprint.pformat(s)

def parse_response(r):
    if r.status_code != 200:
        raise RuntimeError("Request failed with status: "+str(r.status_code)+", "+str(r.content))
    if r.content is not None:
        res = r.json()
        if res['response'] is None:
            raise RuntimeError(res['errors'])
        else:
            return res['response']