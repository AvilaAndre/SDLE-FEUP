// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use rusqlite::{Result};

mod database;
mod controller;
mod state;
pub mod model;
pub mod macros;

use state::{AppState, ServiceAccess};
use tauri::{State, Manager, AppHandle};
use model::*;


fn main(){
    tauri::Builder::default()
    .manage(AppState { db: Default::default() })
    .invoke_handler(tauri::generate_handler![my_custom_command, get_lists, create_list, get_shopping_list, add_item_to_list, update_list_title])
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
fn create_list(app_handle: AppHandle) -> Result<String, String> {
    match app_handle.db(|db| controller::create_list("New List", db)) {
        Err(e) => {
            println!("error creating new list: {e:?}");
            return Err(e.to_string());
        }
        Ok(id) => return Ok(id)
    }
}


#[tauri::command]
fn get_lists(app_handle: AppHandle) -> Result<Vec<ListInfo>, String> {
    match app_handle.db(|db| controller::get_all_lists_info(db)) {
        Err(e) => {
            println!("error getting all lists: {e:?}");
            return Err(e.to_string())
        }
        Ok(items) => return Ok(items)
    }
}

#[tauri::command]
fn get_shopping_list(app_handle: AppHandle, id: String) -> Result<ShoppingListData, String> {
    match app_handle.db(|db| controller::get_list(db, id)) {
        Err(e) => {
            println!("error getting a list: {e:?}");
            return Err(e.to_string())
        }
        Ok(list) => return Ok(list)
    }
}

#[allow(non_snake_case)]
#[tauri::command]
fn add_item_to_list(app_handle: AppHandle, listId: String, name: &str, qtd: i32) -> bool {
    match app_handle.db(|db| controller::add_item_to_list(listId, name, qtd, db)) {
        Err(e) => {
            println!("error creating list item: {e:?}");
            return false;
        }
        Ok(success) => return success,
        
    }
}

#[allow(non_snake_case)]
#[tauri::command]
fn update_list_title(app_handle: AppHandle, listId: String, title: &str) -> bool {
    match app_handle.db(|db| controller::update_list_title(listId, title, db)) {
        Err(e) => {
            println!("error creating list item: {e:?}");
            return false;
        }
        Ok(success) => return success,
    }
}