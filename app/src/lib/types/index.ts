export type ListInfo = {
    list_id: number;
    title: string;
    share_id: String | undefined;
    shared: Boolean;
};

export type ListItemInfo = {
    id: number;
    list_id: number;
    name: string;
    qtd: number;
};

export type ShoppingListData = {
    list_info: ListInfo;
    items: ListItemInfo[];
};

export type TabInfo = {
    title: string;
    ref: string;
    deletable: boolean;
    selected: boolean;
};

export type TabsManager = {
    activeTab: TabInfo;
    values: TabInfo[];
};
