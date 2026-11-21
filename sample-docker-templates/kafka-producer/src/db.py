import src.config as conf
from src.utils.app_db import DatabaseSession

def getDBSession(db_name):
    db = DatabaseSession(
                user=conf.db_user, 
                pwd=conf.db_pass, 
                host=conf.db_host, 
                port=conf.db_port, 
                db_name=db_name)
    return db

def getVIDsFromDB():
    db_conn = getDBSession('devtron_idmap')
    try:
        query = conf.get_vid_query
        vids = db_conn.fetch_all(query)
        return vids
    except Exception as ex:
        raise Exception(f"while getting VIDs... {ex}")
    finally:
        db_conn.close()

def getSaltFromDB(db_conn, modulo):
    try:
        query = """
        select salt from idrepo.uin_hash_salt where id=%s;
        """
        params = [modulo]
        return db_conn.fetch_one(query, params)
    except Exception as ex:
        raise Exception(f"salt not found for modulo {modulo} - {ex}")
      
def getRIDsFromDB(db_conn, mod_uin_hash):
    try:
        query = """
        select reg_id as rid from idrepo.uin where uin_hash=%s;
        """
        params = [mod_uin_hash]
        return db_conn.fetch_one(query, params)
    except Exception as ex:
        raise Exception(f"RID not found for for mod_uin_hash {mod_uin_hash} - {ex}")

def isCredentialTransactionValid(db_conn, where_cond, params):
    try:
        if len(where_cond) > 0:
            query = f"select id from credential.credential_transaction where {where_cond}"
            query = query.format(**params)
            result = db_conn.fetch_one(query)
            if result is not None:
                return False
        return True
    except Exception as ex:
        raise Exception(f"Exception while verify credential transaction - {ex}")
