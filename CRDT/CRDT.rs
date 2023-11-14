use uuid::Uuid;
use std::comp;

mod crdt {

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


        ADD-Wins:

        Σ = P(T × V) × P(T)
        σ
        0
        i = {}, {}
        applyi
        ((add, v),(s,t)) = s ∪ {(utag(), v)},t
        applyi
        ((rmv, v),(s,t)) = s,t ∪ {u | (u, v) ∈ s}
        evali(rd,s) = {v | (u, v) ∈ s ∧ u 6∈ t}
        mergei
        ((s,t),(s
        0
        ,t
        0
        )) = s ∪ s
        0
        ,t ∪ t
        0
    
    
    */
    use std::collections::{ HashMapt, HashSet};

    use uuid::Uuid;


    #[derive(Clone)]
    struct PNCounter {
        positive_count: i32,
        negative_count: i32,
    }

    impl PNCounter{
        fn new() -> Self {
            PNCounter {
                positive_count: 0,
                negative_count: 0,
            }
        }

        fn increment(&mut self, ammount: i32){
            self.positive_count += ammount;
        }

        fn decrement(&mut self, ammount: i32){
            self.negative_count += ammount;
        }

        fn get_count(&self) -> i32 {
            self.positive_count - self.negative_count;
        }

        fn compare(&self, inc_pn_counter) -> Bool{
            self.positive_count <= inc_pn_counter.positive_count && self.negative_count <= inc_pn_counter.negative_count
        }
        // merge function perserving: commutative, associative, and idempotent.
        fn merge(&mut self,  inc_pn_counter: &PNCounter){
            self.positive_count = max(self.positive_count, pncounter.positive_count);
            self.negative_count = max(self.negative_count, pncounter.negative_count);
        }
    }



    #[derive(Clone)]
    struct Item {
        id: Uuid,
        name: String,

        quantity_counter: PNCounter, // Será a quantidade tendo em conta os: increments and decrements

    }

    impl Item {
        fn new(id: Uuid, name: String) -> Self{
            Item {
                id,
                name,
                quantity_counter: PNCounter::new(),
            

            }

        }

        fn increment_quantity(&mut self, increment: i32){
                self.quantity_counter.increment(increment);
            }
        fn decrement_quantity(&mut self, decrement: i32){
                self.quantity_counter.decrement(decrement);
            }

        fn get_quantity(&self) -> i32{
            self.quantity_counter.get_counte(;)
        }
        //Merge current item quantity with other item
        fn merge(&mut self, incoming_item: &Item) {
            if self.id == item.id{
                self.quantity_counter.merge(&incoming_item.quantity_counter);
            }
        }
    }


    #[derive(Clone)]
    struct ShoppingList {
        id: Uuid,
        items: HashMap<Uuid, Item>,
        removed_items: HashSet<Uuid>,
        


    }
    //TODO: Deal with add/remove between lists: choose type of state-CRDT eg: Add-Wins Set?
    impl ShoppingList {
        fn new(id: Uuid) -> Self {
            ShoppingList {
                id,
                items: HashMap::new(),
                removed_items: HashSet::new()
            }


        }
        // Here we assure idempotency for add items: "add item + add item = add item"
        fn add_item(&mut self, incoming_item: Item) {
            self.items.insert(incoming_item.id, incoming_item);
            self.removed_items.remove(&incoming_item.id); // Ensuring add wins over remove
        }
        fn remove_item(&mut self, item_id: Uuid){
               
            self.items.remove(&item_id);
            
            self.removed_items.insert(item_id);
        }
        // TODO: Deal with Add/ remove conflits between items of the lists

        fn merge(&mut self, inc_shopping_list: &ShoppingList){
            if self.id == inc_shopping_list.id{
                for inc_item in inc_shopping_list.items.values(){
                    if !inc_shopping_list.remove_items.contains(&inc_item.id){
                        self.items.insert(inc_item.id, inc_item.clone());
                    }
                }
            }
           
            }

            // Here we merge Merge removed items
            for removed_id in &inc_shopping_list.remove_items{
                if !self.items.contains_key(removed_id){
                    self.remove_items.insert(*removed_id);
                }
            }

            //TODO: How to deal with with conflits about added/removed items from self.list and incomming_list
            //Server side ?!?: Added item with quantity always wins: only when in theory all users of a shared list have removed a certain item, the remove will be done: by default nothing will be merged, because no one have that item
            // even if only one user continue to have added item with quantity, the final state wil be: list with that item
        }
            // Here we verify if any items marked for removal are indeed removed from the active( our items) itemslist
            self.items.retain(|id, _| !self.tombstones.contains(id));

    }

    // TODO: finish this and test the above code
    #[derive(Clone)]
    struct ShoppingListsCRDT {
        lists: HashMap<Uuid, List>,
    }


    #[derive(Clone)]
    impl ShoppingListsCRDT{

        fn new() -> Self {
            ShoppingListsCRDT {
                lists: HashMap::new(),
        
            }


 
        }
       
        fn add_list(% mut self, list_id: Uuid){


        }
        
        fn remove_list(&mut self, list_id: Uuid){

        }


    }
}
    //TODO: How to deal with just send state of the lists that were modified ad replicate/merge across all shared lists