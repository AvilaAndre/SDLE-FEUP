export type ListInfo = {
    list_id: string;
    title: string;
    shared: Boolean;
};

export type AWSet = {
    context: [[]];
    state: [[]];
};

export type CRDT = {
    awset: AWSet;
    items: Object;
    node_id: string;
};

export type CRDTShoppingListData = {
    list_info: ListInfo;
    crdt: CRDT;
    items_checked: Object;
};

export type ListItemInfo = {
    name: string;
    qtd: number;
    checked: boolean;
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
