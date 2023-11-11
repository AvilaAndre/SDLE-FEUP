export type ListInfo = {
    list_id: number;
    title: String;
    share_id: String | undefined;
    shared: Boolean;
};

export type ListItemInfo = {
    id: number;
    list_id: number;
    name: String;
    qtd: number;
};

export type ShoppingListData = {
    list_info: ListInfo;
    items: ListItemInfo[];
};
