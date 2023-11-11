#[derive(serde::Serialize, Debug)]
pub struct ListInfo {
    pub list_id: i32,
    pub title: String,
    pub share_id: Option<String>,
    pub shared: bool,
}

#[derive(serde::Serialize, Debug)]
pub struct ListItemInfo {
    pub id : i32,
    pub list_id : i32,
    pub name : String,
    pub qtd : i32,
}


#[derive(serde::Serialize, Debug)]
pub struct ShoppingListData {
    pub list_info : ListInfo,
    pub items : Vec<ListItemInfo>,
}