use tauri::{AppHandle, Manager, State};
use unqlite::UnQLite;

pub struct AppState {
    pub db: std::sync::Mutex<Option<UnQLite>>,
    pub address: std::sync::Mutex<Option<String>>,
}

pub trait ServiceAccess {
    fn db<F, TResult>(&self, operation: F) -> TResult
    where
        F: FnOnce(&UnQLite) -> TResult;

    fn db_mut<F, TResult>(&self, operation: F) -> TResult
    where
        F: FnOnce(&mut UnQLite) -> TResult;
}

impl ServiceAccess for AppHandle {
    fn db<F, TResult>(&self, operation: F) -> TResult
    where
        F: FnOnce(&UnQLite) -> TResult,
    {
        let app_state: State<AppState> = self.state();
        let db_connection_guard = app_state.db.lock().unwrap();
        let db = db_connection_guard.as_ref().unwrap();

        operation(db)
    }

    fn db_mut<F, TResult>(&self, operation: F) -> TResult
    where
        F: FnOnce(&mut UnQLite) -> TResult,
    {
        let app_state: State<AppState> = self.state();
        let mut db_connection_guard = app_state.db.lock().unwrap();
        let db = db_connection_guard.as_mut().unwrap();

        operation(db)
    }
}

pub trait ServerAddressRecord {
    fn get_server_address(&self) -> String;
    fn set_server_address(&self, new_address: String);
}

impl ServerAddressRecord for AppHandle {
    fn get_server_address(&self) -> String {
        let app_state: State<AppState> = self.state();

        let address: String = app_state.address.lock().unwrap().as_ref().unwrap().clone();

        return address;
    }

    fn set_server_address(&self, new_address: String) {
        let app_state: State<AppState> = self.state();

        *app_state.address.lock().unwrap() = Some(new_address);
    }
}
