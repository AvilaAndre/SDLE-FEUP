use crate::unwrap_or_return_with;
use crate::crdt::crdt::crdt::{ShoppingList};
use std::collections::HashMap;
#[derive(serde::Serialize, serde::Deserialize, Debug)]
pub struct ListInfo {
    pub list_id: String,
    pub title: String,
    pub shared: bool,
   
}


#[derive(serde::Serialize, serde::Deserialize, Debug)]
pub struct ShoppingListData {
    pub list_info: ListInfo,
    //TODO: 
    //for items a user have marked as complete and or quantity goes from positive number to 0 but other users can continue add quantities to the item
    // Quantities will continue to be incremented/decremented, but if marked complete, for that user the frontend display complete with the new quantities from other users: a risk on the item name-quantity or just a marked checkbox
    pub items_checked: HashMap<String, bool>, // item name -> checked or not checked
    pub crdt : ShoppingList
}

pub trait Serialize {
    fn serialize_to_string(&self) -> Result<String, &'static str>;

    fn deserialize_from_slice(value: Vec<u8>) -> Result<ShoppingListData, &'static str>;
}

impl Serialize for ShoppingListData {
    fn serialize_to_string(&self) -> Result<String, &'static str> {
        return match serde_json::to_string(self) {
            Ok(value) => Ok(value),
            Err(_) => return Err("Failed to serialize new data"),
        };
    }

    fn deserialize_from_slice(value: Vec<u8>) -> Result<ShoppingListData, &'static str> {
        return Ok(unwrap_or_return_with!(
            serde_json::from_slice::<ShoppingListData>(&value),
            Err("Failed to deserialize")
        ));
    }
}
