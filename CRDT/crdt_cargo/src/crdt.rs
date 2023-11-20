pub mod crdt {
    use uuid::Uuid;
    use std::collections::HashSet;
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
    // Only grow Counter for context

    #[derive(Clone,Debug)]
    pub struct GCounter {
            positive_count: i32,
            
        }

    impl GCounter{
       pub fn new() -> Self {
            GCounter {
                positive_count: 0,
            }
        }
        pub fn get_positive_count(&self) -> i32 {
            self.positive_count
        }

        
        pub fn increment(&mut self, ammount: i32){
            self.positive_count += ammount;
        }

        

        pub fn get_count(&self) -> i32 {
            self.positive_count 
        }

        pub fn compare(&self, inc_pn_counter: &GCounter) -> bool{
            self.get_positive_count() <= inc_pn_counter.get_positive_count() 
        }
        // merge function perserving: commutative, associative, and idempotent.
        pub fn merge(&mut self, inc_pn_counter: &GCounter) {
            self.positive_count = std::cmp::max(self.positive_count, inc_pn_counter.positive_count);
        }
    }

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

    // // Arranjar estratégias de compressão para os states dos CRDTs !!! Passamos o estado, com o tempo isto vai acumular muita informação


    // Formulation:
    // List of (Items(Name,quanity: PN-Counter), (Nodeid/clientId, only grow counter)

    #[derive(Clone, Debug)]
    pub struct AWSet {
        pub state: HashSet<(String, Uuid, i32)>, // Set of tuples (Item, NodeId, Counter)
        pub context: HashSet<(Uuid, i32)>, // Set of tuples (NodeId, Counter)
    }
    impl AWSet {
        pub fn new() -> Self {
            AWSet{
                state: HashSet::new(),
                context: HashSet::new(), 
            }
        }
        
        // get elements ( Items) of AWSet with corresponding state and context
        pub fn elements(&self) -> String {

        }
        // Check if given element (Item) with corresponding state and context, exist on AWSet
        pub fn contains(item_name: String) -> bool {

        }

        
        //TODO: GCounter ou pertence ao AWSet, ou não é necessário
        // addi(e,(s, c)) -> e= name of created item, s= state, c = context
        // rmvi(e,(s, c)) -> e= name of created item, s= state, c = context
        //maxi(c) = find the max on context set
        // nexti(c) = create the next (NodeId, Counter) taken into account the existing ones on context = (NodeId, Counter) set
        //
        pub fn max_i(&self, node_id: Uuid) -> i32 {
            self.context.iter()
                .filter(|(uuid, _)| *uuid == node_id)
                .map(|(_, counter)| *counter)
                .max()
                .unwrap_or(0)//if there is no (NodeId, Counter) tuple, max_i returns 0, so nex_i can generate tuples(node_id, 0+1( Counter)), when there is no pair on the context
        }
    
        pub fn next_i(&self, node_id: Uuid) -> (Uuid, i32) {
            (node_id, self.max_i(node_id) + 1) 
        }

        // addd new item or increment/decrement existing item: +x increment, -x decrement
        pub fn add_i(&mut self, item_name: String, node_id: Uuid, quantity_change: i32) -> Item{
            let next_context = Self::next_i(&self, node_id);
            self.context.insert(next_context.clone()); // c ∪ {d}

            let existing_item = self.state.iter().find(|(name, _,_)| *name == item_name);

            match existing_item {

                Some((name, _id, _counter)) => {
                    let mut updated_item = Item::new(name.clone());
                    
                    if quantity_change < 0 {
                        let dec_quant_change = -1 * quantity_change;
                        updated_item.decrement_quantity(dec_quant_change);

                    }else if quantity_change > 0{
                        updated_item.increment_quantity(quantity_change);
                    }
                    // just update item without quantity and or return updated_item change with updated next_i
                    self.state.replace((updated_item.name.clone(), next_context.0, next_context.1 )); // s ∪ {(e, d)}
                    return updated_item;

                    


                }

                None => {
                    let mut new_item = Item::new(item_name);
                    if quantity_change < 0 {
                        let dec_quant_change = -1 * quantity_change;
                        new_item.decrement_quantity(dec_quant_change);

                    }else if quantity_change > 0{
                        new_item.increment_quantity(quantity_change);
                    }
                    // just add the item on state with updated next_i
                    self.state.insert((new_item.name.clone(), next_context.0,next_context.1));
                    return new_item;
                }

                

            }
        }
        //Here we remove all tuples from state, that have the item_name
        pub fn rmv_i(&mut self, item_name: String) {
            self.state.retain(|(name,_,_)| *name != item_name);
            
        }

        pub fn filter(&self, inc_awset: &AWSet) -> HashSet<(String, Uuid, i32)> {
            self.state.iter()
            .filter(|(_name, node_id, counter)| {
                !inc_awset.context.iter().any(|(inc_node_id,inc_counter)|{
                    node_id == inc_node_id && counter < inc_counter 
                })
            })
            .cloned()
            .collect()
        }

        pub fn merge(&mut self, inc_awset: &AWSet){
            //Intersection between states of two AWSets
            let states_intersection = self.state.intersection(&inc_awset.state).cloned().collect();
            // Union of filter(s,c') U f(s',c)
            let filter_state_1: HashSet<_> = self.filter(&inc_awset);
            let filter_state_2: HashSet<_> = inc_awset.filter(&self);

            let union_12: HashSet<_> = filter_state_1.union(&filter_state_2).cloned().collect();
            let final_merge: HashSet<_> = union_12.union(&states_intersection).cloned().collect();

            
            // Union of contexts
            let final_context: HashSet<_> = self.context.union(&inc_awset.context).cloned().collect();
            
            // ending merge
            self.state = final_merge;
            self.context = final_context;
        }
    }
    

    // TODO: ShoppingList
    // #[derive(Clone, Debug)]
    // pub struct ShoppingList{
    //     items: HashMap<String, Item>,
    //     awset: AWSet,
    // }
    // impl ShoppingList{
    //     pub fn new() -> Self{
    //         ShoppingList{
    //             items: HashMap::new(), 
    //             awset: AWSet::new(),}
    //     }
    // }

    //add needs to use AWSet, and in the final update/insert one the items atribute
    
}
    

    //TODO: How to deal with just send state of the lists that were modified ad replicate/merge across all shared lists
    #[cfg(test)]
    pub mod tests {
        use crate::crdt::crdt::*;
        use uuid::Uuid;
        use std::collections::HashSet;
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


    //Test AWSet

   

    #[test]
    fn test_awset_new() {
        let awset = AWSet::new();
        assert!(awset.state.is_empty());
        assert!(awset.context.is_empty());
    }

    #[test]
    fn test_max_i() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        awset.context.insert((node_id, 1));
        awset.context.insert((node_id, 3));
        awset.context.insert((node_id, 2));

        assert_eq!(awset.max_i(node_id), 3);
    }

    #[test]
    fn test_next_i() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        awset.context.insert((node_id, 1));
        awset.context.insert((node_id, 2));

        let next = awset.next_i(node_id);
        assert_eq!(next, (node_id, 3));
    }


    #[test]
    fn test_context_with_multiple_nodes() {
        let mut awset = AWSet::new();
        let node_id1 = Uuid::new_v4();
        let node_id2 = Uuid::new_v4();

        
        awset.context.insert((node_id1, 1));
        awset.context.insert((node_id1, 2));
        awset.context.insert((node_id2, 1));
        awset.context.insert((node_id2, 3));

        // Test max_i for different nodes
        assert_eq!(awset.max_i(node_id1), 2);
        assert_eq!(awset.max_i(node_id2), 3);

        // Test next_i for different nodes
        let next1 = awset.next_i(node_id1);
        let next2 = awset.next_i(node_id2);
        assert_eq!(next1, (node_id1, 3));
        assert_eq!(next2, (node_id2, 4));
    }

    #[test]
    fn test_add_new_item() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();
        let quantity_change = 5;

        let item = awset.add_i(item_name.clone(), node_id, quantity_change);
        
        assert_eq!(item.get_name(), "apple");
        assert_eq!(item.get_quantity(), 5);
        assert!(awset.state.contains(&(item_name, node_id, 1)));
        assert!(awset.context.contains(&(node_id, 1)));
    }

    #[test]
    fn test_increment_existing_item() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();
        awset.state.insert((item_name.clone(), node_id, 1));
        awset.context.insert((node_id, 1));

        let item = awset.add_i(item_name.clone(), node_id, 3);
        
        assert_eq!(item.get_name(), "apple");
        assert_eq!(item.get_quantity(), 3);
        assert!(awset.state.contains(&(item_name, node_id, 2)));
        assert!(awset.context.contains(&(node_id, 2)));
    }

    #[test]
    fn test_decrement_existing_item() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();
        awset.state.insert((item_name.clone(), node_id, 1));
        awset.context.insert((node_id, 1));

        let item = awset.add_i(item_name.clone(), node_id, -2);
        
        assert_eq!(item.get_name(), "apple");
        assert_eq!(item.get_quantity(), -2);
        assert!(awset.state.contains(&(item_name, node_id, 2)));
        assert!(awset.context.contains(&(node_id, 2)));
    }

    fn test_add_i() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();

       
        awset.add_i(item_name.clone(), node_id, 1);
        assert_eq!(awset.state.contains(&(item_name.clone(), node_id, 1)), true);
    }

    // Unit tests for rmv_i
    #[test]
    fn test_rmv_i_existing_item() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();

        
        awset.add_i(item_name.clone(), node_id, 1);
        awset.rmv_i(item_name.clone());

        assert_eq!(awset.state.contains(&(item_name, node_id, 1)), false);
    }

    
    #[test]
    fn test_rmv_i_non_existent_item() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();
        let non_existent_item = "banana".to_string();

        
        awset.add_i(item_name.clone(), node_id, 1);
        awset.rmv_i(non_existent_item);

        // Original item should still exist
        assert_eq!(awset.state.contains(&(item_name, node_id, 1)), true);
    }

    
    #[test]
    fn test_rmv_i_context_unchanged() {
        let mut awset = AWSet::new();
        let node_id = Uuid::new_v4();
        let item_name = "apple".to_string();

        
        awset.add_i(item_name.clone(), node_id, 1);
        let context_before_removal = awset.context.clone();
        awset.rmv_i(item_name);

        // Context should remain the same
        assert_eq!(awset.context, context_before_removal);
    }

    #[test]
    fn test_filter_function() {
        let node_id_1 = Uuid::new_v4();
        let node_id_2 = Uuid::new_v4();
        let mut awset_1 = AWSet::new();
        let mut awset_2 = AWSet::new();

        // Setup initial states for both sets
        awset_1.state.insert(("apple".to_string(), node_id_1, 1));
        awset_1.context.insert((node_id_1, 2));

        awset_2.state.insert(("banana".to_string(), node_id_2, 1));
        awset_2.context.insert((node_id_2, 2));

        // Expected result after filtering awset_1 against awset_2: Mock
        let mut expected_state: HashSet<(String, Uuid, i32)> = HashSet::new(); // Assuming "apple" should not be in the filtered state
        expected_state.insert(("apple".to_string(), node_id_1, 1));
        
        let filtered_state = awset_1.filter(&awset_2);

        // Check that the filtered state matches the expected state
        assert_eq!(filtered_state, expected_state);
    }

    //Test merge

    #[test]
    fn test_merge_with_overlap() {
        let mut awset1 = AWSet::new();
        let mut awset2 = AWSet::new();

        let node_id1 = Uuid::new_v4();
        let node_id2 = Uuid::new_v4();
        let counter1 = 1;
        let counter2 = 2;

        
        awset1.state.insert(("apple".to_string(), node_id1, counter1));
        awset2.state.insert(("apple".to_string(), node_id2, counter2));

        // Merging should result in a set that contains both items
        awset1.merge(&awset2);
        assert_eq!(awset1.state.len(), 2);
    }

    #[test]
    fn test_merge_with_distinct_items() {
        let mut awset1 = AWSet::new();
        let mut awset2 = AWSet::new();

        let node_id1 = Uuid::new_v4();
        let node_id2 = Uuid::new_v4();
        let counter1 = 1;
        let counter2 = 2;

       
        awset1.state.insert(("apple".to_string(), node_id1, counter1));
        awset2.state.insert(("banana".to_string(), node_id2, counter2));

        // Merging should result in a set that contains both items
        awset1.merge(&awset2);
        assert_eq!(awset1.state.len(), 2);
    }

    #[test]
    fn test_merge_with_unique_items() {
        let mut awset1 = AWSet::new();
        let awset2 = AWSet::new();

        let node_id = Uuid::new_v4();
        let counter = 1;

        
        awset1.state.insert(("apple".to_string(), node_id, counter));

        // Merging with an empty set should not change the first set
        awset1.merge(&awset2);
        assert_eq!(awset1.state.len(), 1);
        assert!(awset1.state.contains(&("apple".to_string(), node_id, counter)));
    }

    #[test]
    fn test_merge_with_empty_sets() {
        let mut awset1 = AWSet::new();
        let awset2 = AWSet::new();

        // Merging two empty sets should result in an empty set
        awset1.merge(&awset2);
        assert!(awset1.state.is_empty());
    }


}

#[cfg(test)]
mod integration_tests {
    use crate::crdt::crdt::*;
    use uuid::Uuid;


    #[test]
    fn test_add_merge() {
        let mut awset1 = AWSet::new();
        let mut awset2 = AWSet::new();
        let node_id1 = Uuid::new_v4();
        let node_id2 = Uuid::new_v4();

        
        awset1.add_i("apple".to_string(), node_id1, 10);
        awset2.add_i("banana".to_string(), node_id2, 5);

        
        awset1.merge(&awset2);

       // Both id's have added, so the corresponding states are the above
        assert!(awset1.state.contains(&("apple".to_string(), node_id1, 1)));
        assert!(awset1.state.contains(&("banana".to_string(), node_id2, 1)));
        assert!(awset1.context.contains(&( node_id1, 1)));
        assert!(awset1.context.contains(&( node_id2, 1)));
    }

    #[test]
    fn test_remove_merge() {
        let mut awset1 = AWSet::new();
        let mut awset2 = AWSet::new();
        let node_id1 = Uuid::new_v4();
        let node_id2 = Uuid::new_v4();

        
        awset1.add_i("apple".to_string(), node_id1, 10);
        awset1.add_i("apple".to_string(), node_id1, 10);
        awset2.add_i("apple".to_string(), node_id2, 10);
        awset2.rmv_i("apple".to_string());

     
        awset1.merge(&awset2);

        //In the merge, remove needs to stay because id1 have a greater causal context ( Counter =2)
        //So Add wins policy is used
        assert!(awset1.state.contains(&("apple".to_string(), node_id1, 2)));
        assert!(awset1.context.contains(&( node_id1, 2)));
        assert!(!awset2.state.contains(&("apple".to_string(), node_id2, 1)));
        assert!(awset2.context.contains(&(node_id2, 1)));
    }

    #[test]
    fn test_add_remove_same_item_merge() {
        let mut awset1 = AWSet::new();
        let mut awset2 = AWSet::new();
        let node_id1 = Uuid::new_v4();
        let node_id2 = Uuid::new_v4();

       
        awset1.add_i("apple".to_string(), node_id1, 10);
        awset1.rmv_i("apple".to_string());
        awset2.add_i("apple".to_string(), node_id2, 10);
        awset2.rmv_i("apple".to_string());

       
        awset1.merge(&awset2);
        //Both concurrently want to remove apple, remove is done
        assert!(!awset1.state.contains(&("apple".to_string(), node_id1, 1)));
        assert!(!awset1.state.contains(&("apple".to_string(), node_id2, 1)));
        assert!(awset1.context.contains(&( node_id1, 1)));
        assert!(awset2.context.contains(&( node_id2, 1)));
    }


    #[test]
    fn test_remove_multiple_add_merge() {
        let mut awset1 = AWSet::new();
        let mut awset2 = AWSet::new();
        let node_id1 = Uuid::new_v4();
        let node_id2 = Uuid::new_v4();

        
        awset1.add_i("apple".to_string(), node_id1, 10);
        awset1.add_i("apple".to_string(), node_id1, 10);
        awset2.add_i("apple".to_string(), node_id2, 10);
        awset2.add_i("apple".to_string(), node_id2, 10);
        awset2.rmv_i("apple".to_string());

     
        awset1.merge(&awset2);

        // Apple needs to be removed, nodes id1/id2 have the same causal context (Counter =2)
        // It's like we only have one remove, without concurrent add
        assert!(awset1.state.contains(&("apple".to_string(), node_id1, 2)));
        assert!(awset1.context.contains(&( node_id1, 2)));
        assert!(!awset2.state.contains(&("apple".to_string(), node_id2, 2)));
        assert!(awset2.context.contains(&(node_id2, 2)));
    }

    #[test]
    fn test_remove_multiple_add_merge2() {
        let mut awset1 = AWSet::new();
        let mut awset2 = AWSet::new();
        let node_id1 = Uuid::new_v4();
        let node_id2 = Uuid::new_v4();

        
        awset1.add_i("apple".to_string(), node_id1, 10);
        awset1.add_i("apple".to_string(), node_id1, 10);
        awset2.add_i("apple".to_string(), node_id2, 10);
        awset2.add_i("apple".to_string(), node_id2, 10);
        awset2.add_i("apple".to_string(), node_id2, 10);
        awset2.rmv_i("apple".to_string());

     
        awset1.merge(&awset2);

        // Apple needs to be removed, nodes id2 have a greater causal context (Counter =2)
        // It's like id2, is the last user to add apples, and now he wants to remove without concurrents adds from other users
        assert!(awset1.state.contains(&("apple".to_string(), node_id1, 2)));
        assert!(awset1.context.contains(&( node_id1, 2)));
        assert!(!awset2.state.contains(&("apple".to_string(), node_id2, 3)));
        assert!(awset2.context.contains(&(node_id2, 3)));
    }
}