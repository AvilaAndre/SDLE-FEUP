// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use rusqlite::Result;
use uuid::Uuid;
mod controller;
pub mod crdt;
mod database;
pub mod macros;
pub mod model;
mod state;

use model::*;
use state::{AppState, ServiceAccess};
use tauri::{AppHandle, Manager, State};

fn main() {
    tauri::Builder::default()
        .manage(AppState {
            db: Default::default(),
        }) //TODO: add the user new functions
        .invoke_handler(tauri::generate_handler![
            my_custom_command,
            create_user,
            get_user,
            update_user_info,
            get_lists,
            create_list,
            get_shopping_list,
            add_item_to_list,
            update_list_title,
            update_list_item,
            publish_list
        ])
        .setup(|app| {
            let handle = app.handle();

            let app_state: State<AppState> = handle.state();
            let db =
                database::initialize_database(&handle).expect("Database initialize should succeed");
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
//node_id deve ser exterior a qualquer lista e o mesmo utilizado em todas as listas criadas no presente e futuro
//TODO: change to receive node_id, and the title from the input of the frontend and the node_id from created user account using the app
#[tauri::command]
fn create_list(app_handle: AppHandle) -> Result<String, String> {
    match app_handle.db(|db| controller::create_list("New List", Uuid::new_v4(), db)) {
        //TODO: create client info to save client name, node_id: Uuid on local database ( possible also in the) persistent information !!!
        Err(e) => {
            println!("error creating new list: {e:?}");
            return Err(e.to_string());
        }
        Ok(id) => return Ok(id),
    }
}

//TODO: add functions to create client or update name, email, age of the client using the app

//TODO: Create client is the only option available before we can use, create, share lists, etc -> after that the client is stored on local database and the app will always open with created user info

#[tauri::command]
fn create_user(app_handle: AppHandle) -> Result<String, String> {
    //TODO: Change this to receive the inf for parameters from the frontend
    match app_handle.db(|db| controller::create_user("Client", 18, "something@gmail.com", db)) {
        Err(e) => {
            println!("error creating new list: {e:?}");
            return Err(e.to_string());
        }
        Ok(id) => return Ok(id),
    }
}

#[tauri::command]
fn get_user(app_handle: AppHandle, node_id: String) -> Result<User, String> {
    match app_handle.db(|db| controller::get_user(node_id, db)) {
        Err(e) => {
            println!("error getting a list: {e:?}");
            return Err(e.to_string());
        }
        Ok(user) => return Ok(user),
    }
}

#[tauri::command]
fn update_user_info(
    app_handle: AppHandle,
    node_id: String,
    name: &str,
    age: u32,
    email: &str,
) -> bool {
    match app_handle.db(|db| controller::update_user_info(node_id, name, age, email, db)) {
        Err(e) => {
            println!("error updating client {e:?}");
            return false;
        }
        Ok(success) => return success,
    }
}

#[tauri::command]
fn get_lists(app_handle: AppHandle) -> Result<Vec<ListInfo>, String> {
    match app_handle.db(|db| controller::get_all_lists_info(db)) {
        Err(e) => {
            println!("error getting all lists: {e:?}");
            return Err(e.to_string());
        }
        Ok(items) => return Ok(items),
    }
}

#[tauri::command]
fn get_shopping_list(app_handle: AppHandle, id: String) -> Result<ShoppingListData, String> {
    match app_handle.db(|db| controller::get_list(db, id)) {
        Err(e) => {
            println!("error getting a list: {e:?}");
            return Err(e.to_string());
        }
        Ok(list) => return Ok(list),
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


//TODO: do the function for item check using the controller

#[allow(non_snake_case)]
#[tauri::command]
fn update_list_item(
    app_handle: AppHandle,
    listId: String,
    listItem: String,
    counter: u32,
    checked: bool,
) -> Result<SimpleListItem, bool> {
    match app_handle.db(|db| controller::update_list_item(listId, listItem, counter, checked, db)) {
        Err(e) => {
            println!("error creating list item: {e:?}");
            return Err(false);
        }
        Ok(success) => return Ok(success),
    }
}

#[allow(non_snake_case)]
#[tauri::command]
fn publish_list(app_handle: AppHandle, listId: String) -> Result<bool, String> {
    return match app_handle.db(|db| controller::publish_list(listId, db)) {
        Ok(value) => Ok(value),
        Err(reason) => Err(reason.to_string()),
    };
}
