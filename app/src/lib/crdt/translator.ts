import type {
    CRDTShoppingListData,
    ListItemInfo,
    ShoppingListData,
} from "$lib/types";

export const crdtToShoppingList = (
    crdt: CRDTShoppingListData
): ShoppingListData => {
    let shoppingList: ShoppingListData = Object();

    shoppingList.list_info = crdt.list_info;
    shoppingList.items = [];

    console.log(crdt.items_checked); // TODO: implement checked

    Object.entries(crdt.crdt.items).forEach(([key, value]) => {
        let item: ListItemInfo = Object();

        item.name = key;
        item.checked = false;
        item.qtd = 0;

        let negativeCount = 0;
        let positiveCount = 0;

        Object.values(value["negative_count"]).forEach(
            (value: number | any) => {
                negativeCount += value;
            }
        );

        Object.values(value["positive_count"]).forEach(
            (value: number | any) => {
                positiveCount += value;
            }
        );

        item.qtd = positiveCount - negativeCount;

        shoppingList.items.push(item);
    });

    return shoppingList;
};
