use std::fs;
use tauri::AppHandle;
use unqlite::{Cursor, Transaction, UnQLite, KV};

use crate::{model::*, unwrap_or_return, unwrap_or_return_with};

/// Initializes the database connection, creating the .db file if needed, and upgrading the database
/// if it's out of date.
pub fn initialize_database(app_handle: &AppHandle) -> Result<UnQLite, &'static str> {
    let app_dir = app_handle
        .path_resolver()
        .app_data_dir()
        .expect("The app data directory should exist.");
    fs::create_dir_all(&app_dir).expect("The app data directory should be created.");
    let unqlite_path = app_dir.join("shopping_list.db");

    println!("database location {:?}", unqlite_path);
    let db = UnQLite::create(unqlite_path.into_os_string().into_string().unwrap());

    Ok(db)
}

pub trait Operation {
    fn has_key(&self, key: String) -> bool;

    fn store(
        &self,
        id: String,
        list: ShoppingListData,
        reason_if_store_fails: &'static str,
    ) -> Result<bool, &'static str>;

    fn store_user(
        &self,
        node_id: String,
        user: User,
        reason_if_store_fails: &'static str,
    ) -> Result<bool, &'static str>;
    fn get_user(&self, node_id: String) -> Result<User, &'static str>;

    fn get_list(&self, id: String) -> Result<ShoppingListData, &'static str>;

    fn get_all_lists_info(&self) -> Result<Vec<ListInfo>, &'static str>;

    fn delete_list(&self, id: String) -> Result<bool, &'static str>;
}

impl Operation for UnQLite {
    fn has_key(&self, key: String) -> bool {
        return self.kv_contains(key);
    }

    fn store(
        &self,
        id: String,
        list: ShoppingListData,
        reason_if_store_fails: &'static str,
    ) -> Result<bool, &'static str> {
        let serialized_list: String = unwrap_or_return!(list.serialize_to_string());

        unwrap_or_return_with!(
            self.kv_store(id, serialized_list),
            Err(reason_if_store_fails)
        );

        let _ = self.commit();
        return Ok(true);
    }

    fn store_user(
        &self,
        node_id: String,
        user: User,
        reason_if_store_fails: &'static str,
    ) -> Result<bool, &'static str> {
        let serialized_user: String = unwrap_or_return!(user.serialize_to_string());

        unwrap_or_return_with!(
            self.kv_store(node_id, serialized_user),
            Err(reason_if_store_fails)
        );

        let _ = self.commit();
        return Ok(true);
    }

    fn get_user(&self, node_id: String) -> Result<User, &'static str> {
        let result: Vec<u8> = unwrap_or_return_with!(
            self.kv_fetch(node_id),
            Err("Failed to find list with the given id")
        );

        return User::deserialize_from_slice(result);
    }

    fn get_list(&self, id: String) -> Result<ShoppingListData, &'static str> {
        //Adapt for the new type of id for the lists
        let result: Vec<u8> = unwrap_or_return_with!(
            self.kv_fetch(id),
            Err("Failed to find list with the given id")
        );

        return ShoppingListData::deserialize_from_slice(result);
    }

    fn get_all_lists_info(&self) -> Result<Vec<ListInfo>, &'static str> {
        let mut lists: Vec<ListInfo> = Vec::new();

        let mut entry = self.first();

        loop {
            if entry.is_none() {
                break;
            }

            let record = entry.expect("valid entry");
            let (_key, value) = record.key_value();

            let obj: ShoppingListData =
                serde_json::from_slice::<ShoppingListData>(&value).expect("Failed to deserialize");

            lists.push(obj.list_info);

            entry = record.next();
        }

        Ok(lists)
    }

    fn delete_list(&self, id: String) -> Result<bool, &'static str> {
        if self.kv_delete(id).is_err() {
            return Err("failed to delete list");
        }
        return Ok(true);
    }
}
