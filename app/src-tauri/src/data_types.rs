#[derive(serde::Serialize, Debug)]
pub struct ListInfo {
    pub title: String,
    pub share_id: Option<String>,
    pub shared: bool,
}