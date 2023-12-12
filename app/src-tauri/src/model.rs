use crate::crdt::crdt::crdt::ShoppingList;
use crate::unwrap_or_return_with;
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
    pub crdt: ShoppingList,
}

pub trait Serialize {
    type Output;
    fn serialize_to_string(&self) -> Result<String, &'static str>;

    fn deserialize_from_slice(value: Vec<u8>) -> Result<Self::Output, &'static str>; //TODO_ implement for User and ShoppingList
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
        return match serde_json::to_string(self) {
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

#[derive(serde::Serialize, serde::Deserialize, Debug)]
pub struct SimpleListItem {
    pub title: String,
    pub counter: u32,
    pub checked: bool,
}

pub struct SimpleShoppingList {
    pub list_id: String,
    pub title: String,
    pub shared: bool,
    pub items: HashMap<String, SimpleListItem>,//TODO: add purchased and needed 
}

pub trait Simplify {
    type Simple;
    fn simplified(&self) -> Self::Simple;
}

impl Simplify for ShoppingListData {
    type Simple = SimpleShoppingList;

    fn simplified(&self) -> Self::Simple {
        let mut items: HashMap<String, SimpleListItem> = HashMap::<String, SimpleListItem>::new();

        for (item_name, counter) in &self.crdt.items {
            let mut checked: bool = false;

            if self.items_checked.contains_key(item_name) {
                checked = *self.items_checked.get(item_name).unwrap()
            }

            let new_item: SimpleListItem = SimpleListItem {
                title: item_name.to_string(),
                counter: counter.get_count(),
                checked,
            };

            items.insert(item_name.to_string(), new_item);
        }

        let simple_shopping_list: SimpleShoppingList = SimpleShoppingList {
            list_id: self.list_info.list_id.to_owned(),
            title: self.list_info.title.to_owned(),
            shared: self.list_info.shared,
            items,
        };

        return simple_shopping_list;
    }
}
