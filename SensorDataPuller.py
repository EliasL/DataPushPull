import mysql.connector
from mysql.connector import errorcode
import datetime
import threading
import urllib.parse
import urllib.request
import json
import getpass
import base64
import os

__location__ = os.path.realpath(
    os.path.join(os.getcwd(), os.path.dirname(__file__)))

from ctypes import c_char_p, c_longlong, cdll, Structure

TTN = cdll.LoadLibrary(__location__ + "/TTNGet.so")

class GoString(Structure):
    _fields_ = [("p", c_char_p), ("n", c_longlong)]
TTN.GetTTNData.argtypes = [GoString, GoString]
TTN.test.argtypes = [GoString]

# SQL functions
def create_table(cursor, name):
    try:
        cursor.execute("CREATE TABLE {} LIKE template;".format(name))
        print("Creating table {}: ".format(name), end='')

    except mysql.connector.Error as err:
        if err.errno == errorcode.ER_TABLE_EXISTS_ERROR:
            pass
            #print("already exists.")
        else:
            print(err.msg)
    else:
        print("OK")


def add_datapoint(cursor, table, value, date):
    cursor.execute("INSERT IGNORE INTO {} ".format(table)+\
                   "(value, date) VALUES (%s, %s)", (value, date))


# Sensor functions
def get_elwatch_data(sensor_id):
        
    url = "https://neuron.el-watch.com/api/sensordata/"
    API_Key = secret["El-watch"]["API-Key"]
    header = {'API-Key': API_Key}
    req = urllib.request.Request(url+sensor_id, headers=header)
    with urllib.request.urlopen(req) as response:
        data = response.read()
    str_data = data.decode('utf-8')
    json_data = json.loads(str_data)
    sensor_value = json_data['sensors'][0]['last_value']
    sensor_date = json_data['sensors'][0]['last_time']

    return sensor_value, sensor_date, sensor_id

def get_telenor_data(sensor_id):
    url = "https://in.nbiot.engineering/devices/{}/data".format(sensor_id)

    password_mgr = urllib.request.HTTPPasswordMgrWithDefaultRealm()

    # Add the username and password.
    # If we knew the realm, we could use it instead of None.
    password_mgr.add_password(None, url, secret["Telenor"]["username"], secret["Telenor"]["password"])

    handler = urllib.request.HTTPBasicAuthHandler(password_mgr)

    # create "opener" (OpenerDirector instance)
    opener = urllib.request.build_opener(handler)

    # use the opener to fetch a URL
    opener.open(url)

    # Install the opener.
    # Now all calls to urllib.request.urlopen use our opener.
    urllib.request.install_opener(opener)
    with urllib.request.urlopen(url) as response:
        data = response.read()
    json_data = json.loads(data.decode("utf-8"))
    value = base64.b64decode(json_data["data"][-1]["payload"]).decode("utf-8")
    #Temporary because of three values
    if ';' in value:
        value = value.split(";")[0] # From "ab;cd;ef" to ab

    date = datetime.datetime.fromtimestamp(int(json_data["data"][-1]["timestamp"] / 1e9))
    date += datetime.timedelta(hours=2) #Because...
    date = str(date)
    id = json_data["data"][-1]["imei"]
    return value, date, id

def get_TTN_data(sensor_id):
    TTN.test(GoString(b"ok", 2))
    print("ok so far")
    id = sensor_id.encode("utf-8")
    api = secret["The Things Network"]["API-Key"].encode("utf-8")
    id = GoString(id, len(id))
    api = GoString(api, len(api))
    data = TTN.GetTTNData(id, api)
    print(data)
    value = "ok"
    date = "not ok"
    return value, date, id

# Data functions
def pull_sensor_data(sensor_ids):
    data = []

    for id in sensor_ids["SensorTypes"]["El-watch"]:

        sensor_data = get_elwatch_data(id)
        data.append(sensor_data)
    
    for id in sensor_ids["SensorTypes"]["Telenor"]:
        sensor_data = get_telenor_data(id)
        data.append(sensor_data)

    for id in sensor_ids["SensorTypes"]["The Things Network"]:
        sensor_data = get_TTN_data(id)
        data.append(sensor_data)

    return data


def store_sensor_data(cursor, data):
    for sensor_value, sensor_date, sensor_id in data:
        table_name = "sensor_" + sensor_id
        create_table(cursor, table_name)
        add_datapoint(cursor, table_name, sensor_value, sensor_date)


class SensorToDatabase:
    def __init__(self, password):
        self.cnx = mysql.connector.connect(user='elias', password=password,
                                           host='46.101.29.167',
                                           database='elwatch')
        self.cursor = self.cnx.cursor()
        self.sensor_ids = {"SensorTypes": {
                            "El-watch": ["20006040", "20006039", "20004700", "20005880",
                                         "20005883", "20004722", "20004936", "20004874"],
                            "Telenor": ["357517080049085"],
                            "The Things Network": ["power_compare"]
                            }
                        }

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.cursor.close()
        self.cnx.close()

    # Data collection
    def collect_data_every_10_seconds(self):
        threading.Timer(10, self.collect_data_every_10_seconds).start()
        start = datetime.datetime.now()
        print("Data transfere begin: " + str(start))
        print("Begin data download...")
        data = pull_sensor_data(self.sensor_ids)
        print("Begin data upload...")
        store_sensor_data(self.cursor, data)
        print("Data stored. Time to complete: " + str(datetime.datetime.now() - start))
        print("\n|                                                |\n")


with open(__location__ + "/secrets.secret") as json_file:  
    secret = json.load(json_file)

if __name__ == "__main__":
    password = "Jegernomer1997" # CHNAGE! getpass.getpass("Password for mysql user elias at 46.101.29.167: ")
    sensorToDatabase = SensorToDatabase(password)
    sensorToDatabase.collect_data_every_10_seconds()