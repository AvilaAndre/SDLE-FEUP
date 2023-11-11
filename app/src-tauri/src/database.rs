// source: https://github.com/RandomEngy/tauri-sqlite/blob/main/src-tauri/src/database.rs


use rusqlite::{Connection, named_params};
use tauri::AppHandle;
use std::fs;

use crate::data_types::{ListInfo};

const CURRENT_DB_VERSION: u32 = 1;

/// Initializes the database connection, creating the .sqlite file if needed, and upgrading the database
/// if it's out of date.
pub fn initialize_database(app_handle: &AppHandle) -> Result<Connection, rusqlite::Error> {
    let app_dir = app_handle.path_resolver().app_data_dir().expect("The app data directory should exist.");
    fs::create_dir_all(&app_dir).expect("The app data directory should be created.");
    let sqlite_path = app_dir.join("shopping_list.sqlite");
    
    println!("{:?}", sqlite_path);
    let mut db = Connection::open(sqlite_path)?;


    let mut user_pragma = db.prepare("PRAGMA user_version")?;
    let existing_user_version: u32 = user_pragma.query_row([], |row| { Ok(row.get(0)?) })?;
    drop(user_pragma);

    upgrade_database_if_needed(&mut db, existing_user_version)?;

    Ok(db)
}

/// Upgrades the database to the current version.
pub fn upgrade_database_if_needed(db: &mut Connection, existing_version: u32) -> Result<(), rusqlite::Error> {
    if true { // TODO: replace true with 'existing_version < CURRENT_DB_VERSION'
        db.pragma_update(None, "journal_mode", "WAL")?;

        let tx = db.transaction()?;

        tx.pragma_update(None, "user_version", CURRENT_DB_VERSION)?;

        tx.execute_batch(
            "CREATE TABLE IF NOT EXISTS shopping_list (
                list_id INTEGER PRIMARY KEY AUTOINCREMENT,
                title TEXT NOT NULL,
                share_id TEXT UNIQUE DEFAULT NULL,
                shared INTEGER NOT NULL
            );
            CREATE TABLE IF NOT EXISTS list_item (
                id INTEGER PRIMARY KEY,
                list_id INTEGER NOT NULL REFERENCES shopping_list(list_id),
                name TEXT NOT NULL,
                qtd INTEGER NOT NULL
            );",
        )?;

        tx.commit()?;
    }

    Ok(())
}


/**
 *  Creates a new list and returns the created id
 * */
pub fn create_list(title: &str, db: &Connection) -> Result<i32, rusqlite::Error> {
    let mut statement = db.prepare("INSERT INTO shopping_list (title, shared) VALUES (@title, @shared)")?;
    statement.execute(named_params! { "@title": title, "@shared": 0 })?;


    statement = db.prepare("select seq from sqlite_sequence where name=\"shopping_list\"")?;
    let mut rows = statement.query([])?;
    while let Some(row) = rows.next()? {
        let new_id: i32 = row.get("seq")?;
        
        return Ok(new_id)
    }

    Ok(-1)
}

pub fn get_all_lists(db: &Connection) -> Result<Vec<ListInfo>, rusqlite::Error> {
    let mut statement = db.prepare("SELECT * FROM shopping_list")?;
    let mut rows = statement.query([])?;
    let mut items: Vec<ListInfo> = Vec::new();
    while let Some(row) = rows.next()? {
        let share_id: Option<String>;
        match row.get("share_id") {
            Ok(value) => share_id = Some(value),
            Err(_) => share_id = None
        }

        let shared: bool;
        match row.get("shared") {
            Ok(0) => shared = false,
            Ok(1) => shared = true,
            Ok(value) => {
                println!("shared column value should be 1 or 0, it is {}. Defaulting to true", value);
                shared = false
            }
            Err(e) => {
                println!("shared column value error: {e:?}");
                shared = false
            }
        }

        let new_item: ListInfo = ListInfo {
            title: row.get("title")?,
            share_id,
            shared
        
        };
    
        items.push(new_item);
    }
    
    Ok(items)
}