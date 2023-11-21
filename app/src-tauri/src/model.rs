use crate::unwrap_or_return_with;

#[derive(serde::Serialize, serde::Deserialize, Debug)]
pub struct ListInfo {
    pub list_id: String,
    pub title: String,
    pub shared: bool,
}

#[derive(serde::Serialize, serde::Deserialize, Debug)]
pub struct ListItemInfo {
    pub id: i32,
    pub name: String,
    pub qtd: i32,
}

#[derive(serde::Serialize, serde::Deserialize, Debug)]
pub struct ShoppingListData {
    pub list_info: ListInfo,
    pub items: Vec<ListItemInfo>,
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
