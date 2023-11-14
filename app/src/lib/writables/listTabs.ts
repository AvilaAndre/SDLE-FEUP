import { goto } from "$app/navigation";
import type { TabInfo } from "$lib/types";
import { writable } from "svelte/store";

export const tabsList = writable([
    {
        title: "My Lists",
        ref: "/",
        deletable: false,
        selected: true,
    },
]);

export const openTab = (title: string, ref: string) => {
    let newTab: TabInfo = {
        title,
        ref,
        deletable: true,
        selected: true,
    };

    tabsList.update((value) => {
        value.forEach((item) => {
            item.selected = false;
        });
        for (let index = 0; index < value.length; index++) {
            const element = value[index];

            if (element.ref === newTab.ref) {
                element.title = newTab.title;
                element.selected = true;
                return value;
            }
        }

        return [...value, newTab];
    });
};

export const closeTab = (ref: string) => {
    tabsList.update((value) => {
        let updatedList: TabInfo[] = [];

        for (let index = 0; index < value.length; index++) {
            const element = value[index];

            if (element.ref === ref && updatedList.length != 0) {
                if (element.selected) {
                    updatedList[0].selected = true;
                    goto("/");
                }
                continue;
            }
            updatedList.push(element);
        }

        return updatedList;
    });
};
