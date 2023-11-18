pub mod crdt {
    use uuid::Uuid;
    
    /*
        To help: PN-Counter for Item of a list

        payload integer[n] P, integer[n] N
            initial [0,0,...,0], [0,0,...,0]
        update increment()
            let g = myId()
            P[g] := P[g] + 1
        update decrement()
            let g = myId()
            N[g] := N[g] + 1
        query value() : integer v
            let v = Σi P[i] - Σi N[i]
        compare (X, Y) : boolean b
            let b = (∀i ∈ [0, n - 1] : X.P[i] ≤ Y.P[i] ∧ ∀i ∈ [0, n - 1] : X.N[i] ≤ Y.N[i])
        merge (X, Y) : payload Z
            let ∀i ∈ [0, n - 1] : Z.P[i] = max(X.P[i], Y.P[i])
            let ∀i ∈ [0, n - 1] : Z.N[i] = max(X.N[i], Y.N[i]
        
    */

    #[derive(Clone,Debug)]
   pub struct PNCounter {
        positive_count: i32,
        negative_count: i32,
    }

    impl PNCounter{
       pub fn new() -> Self {
            PNCounter {
                positive_count: 0,
                negative_count: 0,
            }
        }
        pub fn get_positive_count(&self) -> i32 {
            self.positive_count
        }

        pub fn get_negative_count(&self) -> i32 {
            self.negative_count
        }
        pub fn increment(&mut self, ammount: i32){
            self.positive_count += ammount;
        }

        pub fn decrement(&mut self, ammount: i32){
            self.negative_count += ammount;
        }

        pub fn get_count(&self) -> i32 {
            self.positive_count - self.negative_count
        }

        pub fn compare(&self, inc_pn_counter: &PNCounter) -> bool{
            self.get_positive_count() <= inc_pn_counter.get_positive_count() && self.get_negative_count() <= inc_pn_counter.get_negative_count()
        }
        // merge function perserving: commutative, associative, and idempotent.
        pub fn merge(&mut self, inc_pn_counter: &PNCounter) {
            self.positive_count = std::cmp::max(self.positive_count, inc_pn_counter.positive_count);
            self.negative_count = std::cmp::max(self.negative_count, inc_pn_counter.negative_count);
        }
    }



    #[derive(Clone,Debug)]
    pub struct Item {
        // id: Uuid,
        name: String, // comparar por nome -> ver restrições a garantir

        quantity_counter: PNCounter, // Será a quantidade tendo em conta os: increments and decrements

    }

    impl Item {
        pub fn new( name: String) -> Self{
            Item {
                // id, check this later
                name,
                quantity_counter: PNCounter::new(),
            

            }

        }

        // pub fn get_id(&self) -> uuid::Uuid {
        //     self.id
        // }
    
        pub fn get_name(&self) -> &str {
            &self.name
        }

        pub  fn increment_quantity(&mut self, increment: i32){
                self.quantity_counter.increment(increment);
            }
        pub  fn decrement_quantity(&mut self, decrement: i32){
                self.quantity_counter.decrement(decrement);
            }

        pub  fn get_quantity(&self) -> i32{
            return self.quantity_counter.get_count();
        }
        //Merge current item quantity with other item
        pub  fn merge(&mut self, incoming_item: &Item) {
            if self.name == incoming_item.name{
                self.quantity_counter.merge(&incoming_item.quantity_counter);
            }
        }
    }
}
    // // Arranjar estratégias de compressão para os states dos CRDTs !!! Passamos o estado, com o tempo isto vai acumular muita informação

    // #[derive(Clone)]
    // struct ShoppingList {
    //     id: Uuid,
    //     items: HashMap<String, Item>,
        


    // }

    // TODO: check this later
    // the tuple (item: Item, updated: bool), the updated is used to know what we need to merge or not
    // items: HashMap<item_name: String, (item: Item, updated: bool)> ( To check later) -> pode ser usado fora do CRDT para fazer tracking dos states item mudados, e assim só enviar estes ?!!?
    // //TODO: Deal with add/remove between lists: choose type of state-CRDT eg: Observed Set with Add-Wins strategy? Pode fazer sentido ter clock
    // impl ShoppingList {
    //     fn new(id: Uuid) -> Self {
    //         ShoppingList {
    //             id,
    //             items: HashMap::new(),
    //             removed_items: HashSet::new()
    //         }


    //     }
    //     // Here we assure idempotency for add items: "add item + add item = add item"
    //     fn add_item(&mut self, incoming_item: Item) {
    //         self.items.insert(incoming_item.id, incoming_item);
    //         self.removed_items.remove(&incoming_item.id); // Ensuring add wins over remove
    //     }
    //     fn remove_item(&mut self, item_id: Uuid){
               
    //         self.items.remove(&item_id);
            
    //         self.removed_items.insert(item_id);
    //     }
    //     // TODO: Deal with Add/ remove conflits between items of the lists

    //     fn merge(&mut self, inc_shopping_list: &ShoppingList){
    //         if self.id == inc_shopping_list.id{
    //             for inc_item in inc_shopping_list.items.values(){
    //                 if !inc_shopping_list.remove_items.contains(&inc_item.id){
    //                     self.items.insert(inc_item.id, inc_item.clone());
    //                 }
    //             }
    //         }
           
        

    //         // Here we merge Merge removed items
    //         for removed_id in &inc_shopping_list.remove_items{
    //             if !self.items.contains_key(removed_id){
    //                 self.remove_items.insert(*removed_id);
    //             }
    //         }

    //         //TODO: How to deal with with conflits about added/removed items from self.list and incomming_list
    //         //Server side ?!?: Added item with quantity always wins: only when in theory all users of a shared list have removed a certain item, the remove will be done: by default nothing will be merged, because no one have that item
    //         // even if only one user continue to have added item with quantity, the final state wil be: list with that item
    //     }
    //         // Here we verify if any items marked for removal are indeed removed from the active( our items) itemslist
    //         self.items.retain(|id, _| !self.removed_items.contains(id));

    // }

    // TODO: finish this and test the above code
    // #[derive(Clone)]
    // struct ShoppingListsCRDT {
    //     lists: HashMap<Uuid, List>,
    // }


    // #[derive(Clone)]
    // impl ShoppingListsCRDT{

    //     fn new() -> Self {
    //         ShoppingListsCRDT {
    //             lists: HashMap::new(),
        
    //         }


 
    //     }
       
    //     fn add_list(% mut self, list_id: Uuid){


    //     }
        
    //     fn remove_list(&mut self, list_id: Uuid){

    //     }


    // }

    //TODO: How to deal with just send state of the lists that were modified ad replicate/merge across all shared lists
    #[cfg(test)]
    pub mod tests {
        use crate::crdt::crdt::*;
        // use uuid::Uuid;
        #[test]
        fn test_increment_pncounter() {
            let mut counter = PNCounter::new();
            counter.increment(5);
            assert_eq!(counter.get_positive_count(), 5);
        }
    
        #[test]
        fn test_decrement_pncounter() {
            let mut counter = PNCounter::new();
            counter.decrement(3);
            assert_eq!(counter.get_negative_count(), 3);
        }
    
    
        #[test]
        fn test_get_count_pncounter() {
            let mut counter = PNCounter::new();
            counter.increment(10);
            counter.decrement(4);
            assert_eq!(counter.get_count(), 6);
        }
    
        #[test]
        fn test_pncounter_compare() {
            let mut counter1 = PNCounter::new();
            let mut counter2 = PNCounter::new();
            counter1.increment(5);
            counter2.increment(3);
            assert!(counter2.compare(&counter1));
        }
    
        #[test]
        fn test_pncounter_merge() {
            let mut counter1 = PNCounter::new();
            let mut counter2 = PNCounter::new();
            counter1.increment(10);
            counter2.increment(5);
            counter1.merge(&counter2);
            assert_eq!(counter1.get_count(), 10); // Ensure it takes the max
        }



    // Testing Item CRDT

    #[test]
    fn test_item_creation() {
        // let id = Uuid::new_v4();
        let name = String::from("test_item");
        let item = Item::new(name.clone());

        // assert_eq!(item.get_id(), id);
        assert_eq!(item.get_name(), name);
        assert_eq!(item.get_quantity(), 0);
    }

    #[test]
    fn test_increment_quantity() {
        let mut item = Item::new( String::from("test_item"));
        item.increment_quantity(5);

        assert_eq!(item.get_quantity(), 5);
    }

    #[test]
    fn test_decrement_quantity() {
        let mut item = Item::new( String::from("test_item"));
        item.increment_quantity(10);
        item.decrement_quantity(3);

        assert_eq!(item.get_quantity(), 7);
    }

    #[test]
    fn test_merge_items_same_name() {
        // let id = Uuid::new_v4();
        let mut item1 = Item::new( String::from("test_item1"));
        let mut item2 = Item::new( String::from("test_item1"));

        item1.increment_quantity(5);
        item2.increment_quantity(10);
        item1.merge(&item2);

        assert_eq!(item1.get_quantity(), 10);
    }

    #[test]
    fn test_no_merge_for_different_names() {
        let mut item1 = Item::new(String::from("test_item1"));
        let mut item2 = Item::new(String::from("test_item2"));

        item1.increment_quantity(5);
        item2.increment_quantity(10);
        item1.merge(&item2);

        assert_eq!(item1.get_quantity(), 5);
    }

}