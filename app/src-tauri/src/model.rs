use crate::unwrap_or_return_with;
use crate::crdt::crdt::crdt::{ShoppingList};
use std::collections::HashMap;
//TODO: an user maybe can have also a HashMap with list_Uuid -> list_name: A map of the lists the client have
#[derive(serde::Serialize, serde::Deserialize, Debug)]
pub struct User {
    pub node_id: String,
    pub name: String,
    pub age: u32,
    pub email: String,
}


        

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
    pub items_checked: HashMap<String, bool>, // item name -> checked or not checked( True or False) -> check controller for details
    pub crdt : ShoppingList
}

pub trait Serialize {
    type Output;
    fn serialize_to_string(&self) -> Result<String, &'static str>;

    fn deserialize_from_slice(value: Vec<u8>) -> Result<Self::Output, &'static str>;//TODO_ implement for User and ShoppingList
    
    
    
}

impl Serialize for ShoppingListData {
    type Output = ShoppingListData;

    fn serialize_to_string(&self) -> Result<String, &'static str> {
        return match serde_json::to_string(self) {
            Ok(value) => Ok(value),
            Err(_) => return Err("Failed to serialize new data"),
        };
    }

    fn deserialize_from_slice(value: Vec<u8>) -> Result<Self::Output, &'static str> {
        return Ok(unwrap_or_return_with!(
            serde_json::from_slice::<Self::Output>(&value),
            Err("Failed to deserialize")
        ));
    }
}

impl Serialize for User {

    type Output = User;
    fn serialize_to_string(&self) -> Result<String, &'static str> {
        return match serde_json::to_string(self){
            Ok(value) => Ok(value),
            Err(_) => return Err("Failed to serialize new client data"),
        };


    }

    fn deserialize_from_slice(value: Vec<u8>) -> Result<Self::Output, &'static str> {
        return Ok(unwrap_or_return_with!(
            serde_json::from_slice::<Self::Output>(&value),
            Err("Failed to deserialize")
        ));
    }
}
