// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use rand::{self, Rng, distributions::Uniform, prelude::Distribution};

use rusqlite::{Connection, Result};

mod database;
mod state;
pub mod data_types;

use state::{AppState, ServiceAccess};
use tauri::{State, Manager, AppHandle};
use data_types::{ListInfo};


fn main(){
    tauri::Builder::default()
    .manage(AppState { db: Default::default() })
    .invoke_handler(tauri::generate_handler![my_custom_command, get_mock_data, get_lists, create_list])
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

#[derive(serde::Serialize)]
struct ListData {
    title: String,
    items: Vec<String>,
}

#[tauri::command]
fn get_mock_data() -> Result<ListData, String> {

    let generated_title: String = "Title from Rust!".to_string();
    
    let generated_items: Vec<String> = vec![
        "apples".to_string(),
        "shitakes".to_string(),
        "celery seeds".to_string(),
        "sherry".to_string(),
        "sunflower seeds".to_string(),
        "blackberries".to_string(),
        "passion fruit".to_string(),
        "Goji berry".to_string(),
        "lettuce".to_string(),
        "sweet potatoes".to_string(),
        "capers".to_string(),
        "almond paste".to_string(),
        "tea".to_string(),
        "powdered sugar".to_string(),
        "zinfandel wine".to_string(),
        "rosemary".to_string(),
        "cumin".to_string(),
        "five-spice powder".to_string(),
        "rum".to_string(),
        "wine vinegar".to_string(),
        "brown rice".to_string(),
        "bagels".to_string(),
        "cranberries".to_string(),
        "turnips".to_string(),
        "fennel seeds".to_string(),
        "wild rice".to_string(),
        "olives".to_string(),
        "tomato paste".to_string(),
        "cactus".to_string(),
        "spaghetti squash".to_string()
    ];

    let mut rng = rand::thread_rng();

    let n1: u8 = rng.gen_range(2..30);

    let mut i = 0;

    let mut items: Vec<String> = vec![];

    while i < n1 {
        let throw = Uniform::from(0..generated_items.len()).sample(&mut rng);
        i = i+1;

        items.push(generated_items.get(throw).unwrap().to_string())
    }
    
    Ok(ListData { title: generated_title, items } )
}