#[derive(serde::Serialize, serde::Deserialize, Debug)]
pub struct ListInfo {
    pub list_id: String,
    pub title: String,
    pub shared: bool,
}

#[derive(serde::Serialize, serde::Deserialize, Debug)]
pub struct ListItemInfo {
    pub id : i32,
    pub name : String,
    pub qtd : i32,
}


#[derive(serde::Serialize, serde::Deserialize, Debug)]
pub struct ShoppingListData {
    pub list_info : ListInfo,
    pub items : Vec<ListItemInfo>,
}