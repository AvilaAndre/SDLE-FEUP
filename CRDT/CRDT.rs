mod crdt {
    use std::collections::{ HashMapt, HashSet};

    use uuid::Uuid;

    #[derive(Clone)]
    struct Item {
        id: Uuid,
        quantity: i32, // Será a quantidade tendo em conta os: increments and decrements

    }

    #[derive(Clone)]
    struct List {
        id: Uuid,
        items: HashMap<Uuid, Item>,

        


    }




    #[derive(Clone)]
    struct ShoppingListsCRDT {
        lists: HashMap<Uuid, List>,
    }


    #[derive(Clone)]
    impl ShoppingCRDT{

        fn new() -> Self {
            ShoppingListsCRDT {
                lists: HashMap::new(),
        
            }


 
        }
       
        fn add_list(% mut self, list_id: Uuid){


        }
        
        fn remove_list(&mut self, list_id: Uuid){

        }


        fn add_item(&mut self, list_id: &Uuid, item_id: Uuid, name: String){

        }

        fn remove_item(&mut self, list_id: &Uuid, item_id: Uuid, name: String){

        }

        fn add_item(){


        }

        fn increment_item(&mut self, list_id: &Uuid, item_id: Uuid, ammount: i32){

        }

        fn decrement_item(&mut self, list_id: &Uuid, item_id: Uuid, ammount: i32) {


        }



    }
}