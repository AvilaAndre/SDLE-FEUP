use crate::database::*;
use crate::macros::*;
use crate::model::*;

use uuid::Uuid;

use unqlite::UnQLite;

/**
 *  Creates a new list and returns the created id
 * */
pub fn create_list(title: &str, db: &UnQLite) -> Result<String, &'static str> {
    let mut id = Uuid::new_v4().to_string();

    while db.has_key(id.clone()) {
        id = Uuid::new_v4().to_string();
    }

    let new_list: ShoppingListData = ShoppingListData {
        list_info: ListInfo {
            list_id: id.clone(),
            title: title.to_string(),
            shared: false,
        },
        items: Vec::new(),
    };

    unwrap_or_return!(db.store(id.clone(), new_list, "Failed to store new list"));

    return Ok(id);
}

/**
 * Gets the data of a list
 */
pub fn get_list(db: &UnQLite, id: String) -> Result<ShoppingListData, &'static str> {
    return db.get_list(id);
}

/**
 * Gets information about every list in the database
 */
pub fn get_all_lists_info(db: &UnQLite) -> Result<Vec<ListInfo>, &'static str> {
    return db.get_all_lists_info();
}

/**
 *  Adds new list item to a specified list
 * */
pub fn add_item_to_list(
    list_id: String,
    name: &str,
    qtd: i32,
    db: &UnQLite,
) -> Result<bool, &'static str> {
    let mut list = unwrap_or_return!(db.get_list(list_id.clone()));

    list.items.push(ListItemInfo {
        id: 0,
        name: name.to_string(),
        qtd,
    });

    return Ok(unwrap_or_return!(db.store(
        list_id,
        list,
        "Failed to store updated list"
    )));
}

/**
 *  Updates a specified list's title
 * */
pub fn update_list_title(list_id: String, title: &str, db: &UnQLite) -> Result<bool, &'static str> {
    let mut list: ShoppingListData = unwrap_or_return!(db.get_list(list_id.clone()));

    list.list_info.title = title.to_string();

    return Ok(unwrap_or_return!(db.store(
        list_id,
        list,
        "Failed to store updated list"
    )));
}
