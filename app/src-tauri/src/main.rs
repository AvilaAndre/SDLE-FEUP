// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use rand::{self, Rng, distributions::Uniform, prelude::Distribution};

use rusqlite::{Connection, Result};

mod database;
mod state;
pub mod data_types;

use state::{AppState, ServiceAccess};
use tauri::{State, Manager, AppHandle};
use data_types::*;


fn main(){
    tauri::Builder::default()
    .manage(AppState { db: Default::default() })
    .invoke_handler(tauri::generate_handler![my_custom_command, get_lists, create_list, get_shopping_list, add_item_to_list])
    .setup(|app| {
        let handle = app.handle();

        let app_state: State<AppState> = handle.state();
        let db = database::initialize_database(&handle).expect("Database initialize should succeed");
        *app_state.db.lock().unwrap() = Some(db);

        Ok(())
    })
    .run(tauri::generate_context!())
    .expect("error while running tauri application");
}


#[tauri::command]
fn my_custom_command() {
    println!("I was invoked from JS!");
}

#[tauri::command]
fn create_list(app_handle: AppHandle) -> Result<i32, String> {
    match app_handle.db(|db| database::create_list("New List", db)) {
        Err(e) => {
            println!("error creating new list: {e:?}");
            return Err(e.to_string());
        }
        Ok(id) => return Ok(id)
    }
}


#[tauri::command]
fn get_lists(app_handle: AppHandle) -> Result<Vec<ListInfo>, String> {
    match app_handle.db(|db| database::get_all_lists(db)) {
        Err(e) => {
            println!("error getting all lists: {e:?}");
            return Err(e.to_string())
        }
        Ok(items) => return Ok(items)
    }
}

#[tauri::command]
fn get_shopping_list(app_handle: AppHandle, id: i32) -> Result<ShoppingListData, String> {
    match app_handle.db(|db| database::get_list(id, db)) {
        Err(e) => {
            println!("error getting all lists: {e:?}");
            return Err(e.to_string())
        }
        Ok(list) => match list {
            Some(list_data) => return Ok(list_data),
            None => return Err("Failed to get list data".to_string())
        }
    }
}

#[tauri::command]
fn add_item_to_list(app_handle: AppHandle, listId: i32, name: &str, qtd: i32) -> bool {
    match app_handle.db(|db| database::add_item_to_list(listId, name, qtd, db)) {
        Err(e) => {
            println!("error creating list item: {e:?}");
            return false;
        }
        Ok(success) => return success,
        
    }
}