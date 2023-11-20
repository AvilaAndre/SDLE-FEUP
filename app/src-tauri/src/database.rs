// source: https://github.com/RandomEngy/tauri-sqlite/blob/main/src-tauri/src/database.rs


use uuid::Uuid;

use unqlite::{UnQLite, KV, Cursor, Transaction};
use tauri::AppHandle;
use std::fs;

use crate::data_types::*;

/// Initializes the database connection, creating the .db file if needed, and upgrading the database
/// if it's out of date.
pub fn initialize_database(app_handle: &AppHandle) -> Result<UnQLite, &'static str> {
    let app_dir = app_handle.path_resolver().app_data_dir().expect("The app data directory should exist.");
    fs::create_dir_all(&app_dir).expect("The app data directory should be created.");
    let unqlite_path = app_dir.join("shopping_list.db");
    
    println!("database location {:?}", unqlite_path);
    let db = UnQLite::create(unqlite_path.into_os_string().into_string().unwrap());

    Ok(db)
}

/**
 *  Creates a new list and returns the created id
 * */
pub fn create_list(title: &str, db: &UnQLite) -> Result<String, &'static str> {
    let mut id = Uuid::new_v4().to_string();

    while db.kv_contains(id.clone()) {
        id = Uuid::new_v4().to_string();
    }
    
    let value: ShoppingListData = ShoppingListData {
        list_info: ListInfo { list_id: id.clone(), title: title.to_string(), shared: false },
        items: Vec::new()
    };
    
    let serialized_value =  match serde_json::to_string(&value) {
        Ok(value) => value,
        Err(_) => return Err("Failed to serialize new data")
    };


    match db.kv_store(id.clone(), serialized_value) {
        Err(_e) => return Err("Failed to store new list"),
        Ok(_) => {
            let _ = db.commit(); // TODO: Check if the commit was sucessfull, the rollback is done automatically if not
            return Ok(id)
        }
    }
}

pub fn get_all_lists(db: &UnQLite) -> Result<Vec<ListInfo>, &'static str> {
    let mut items: Vec<ListInfo> = Vec::new();

    let mut entry = db.first();

    loop {
        if entry.is_none() { break; }

        let record = entry.expect("valid entry");
        let (_key, value) = record.key_value();

        let obj: ShoppingListData = serde_json::from_slice::<ShoppingListData>(&value).expect("Failed to deserialize");

        items.push(obj.list_info);

        entry = record.next();
    }

    Ok(items)
}

pub fn get_list(id: String, db: &UnQLite) -> Result<ShoppingListData, &'static str> {
    let result : Vec<u8> = match db.kv_fetch(id) {
        Ok(value) => value,
        Err(_) => return Err("failed to find list with the given id")
    };

    match serde_json::from_slice::<ShoppingListData>(&result) {
        Ok(shopping_list) => return Ok(shopping_list),
        Err(_) => return Err("Failed to deserialize")
    };
}

/**
 *  Adds new list item to a specified list
 * */
pub fn add_item_to_list(list_id: String, name: &str, qtd: i32, db: &UnQLite) -> Result<bool, &'static str> {
    let mut list: ShoppingListData = match get_list(list_id.clone(), db) {
        Ok(value) => value,
        Err(error) => return Err(error)
    };

    list.items.push( ListItemInfo { id: 0, name: name.to_string(), qtd });

    let serialized_value =  match serde_json::to_string(&list) {
        Ok(value) => value,
        Err(_) => return Err("Failed to serialize updated list")
    };

    match db.kv_store(list_id, serialized_value) {
        Err(_e) => return Err("Failed to store updated list"),
        Ok(_) =>  {
            let _ = db.commit();
            return Ok(true)
        }
    }
}

/**
 *  Updates a specified list's title
 * */
pub fn update_list_title(list_id: String, title: &str, db: &UnQLite) -> Result<bool, &'static str> {
    let mut list: ShoppingListData = match get_list(list_id.clone(), db) {
        Ok(value) => value,
        Err(error) => return Err(error)
    };

    list.list_info.title = title.to_string();

    let serialized_value =  match serde_json::to_string(&list) {
        Ok(value) => value,
        Err(_) => return Err("Failed to serialize updated list")
    };

    match db.kv_store(list_id, serialized_value) {
        Err(_e) => return Err("Failed to store updated list"),
        Ok(_) => {
            let _ = db.commit();
            return Ok(true)
        }
    }
}