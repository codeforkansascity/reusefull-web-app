
Vue.component('charity-card', {
  props: ['charity'],
  data() {
    return {
      charityLink: '/charity/'+this.charity.id,
      phoneLink: 'tel:'+this.charity.phone,
      phone: `(${this.charity.phone.substr(0,3)}) ${this.charity.phone.substr(4,3)}-${this.charity.phone.substr(6,4)}`,
      mapLink: 'https://www.google.com/maps/dir/?api=1&destination='+ encodeURIComponent(this.charity.address)
    }
  },
  template: `
  <div class="card shadow p-3 mb-3 bg-white rounded">
  <div class="row">
    <div class="col-auto">
      <a :href="charity.id">
        <img
          :src="charity.logoURL"
          class="img-fluid"
          width="200"
          alt="Org logo"
        />
      </a>
    </div>
    <div class="col">
      <div class="card-bock">
        <h5 class="card-title"><a :href="charityLink">{{charity.name}}</a></h5>
        <div class="card-text">
          <div class="my-1"><strong>Pick-Up:</strong>
            <i class="bi bi-check-circle" style="color: green" v-if="charity.pickup"></i>
            <i class="bi bi-x-circle" style="color: red" v-if="!charity.pickup"></i>
          </div>
          <div class="my-1"><strong>Dropoff:</strong>
            <i class="bi bi-check-circle" style="color: green" v-if="charity.dropoff"></i>
            <i class="bi bi-x-circle" style="color: red" v-if="!charity.dropoff"></i>
          </div>
        </div>
        <div>
          <i class="bi bi-geo-alt-fill"></i>
          <strong>Address:</strong>
          <em>{{charity.address}} </em>
          <a target="_blank" :href="mapLink"><i class="bi bi-geo p-1"></i>Directions</a>
        </div>
        <div>
          <i class="bi bi-telephone-fill"></i>
          <strong>Phone:</strong>
          <a :href="phoneLink">{{phone}}</a>
        </div>
      </div>
    </div>
  </div>
  </div>
    `
});

const app = new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
    errors: [],
    donate: {
      itemTypes: [],
      charityTypes: [],
      anyCharityType: null,
      budget: null,
      pickupDropoff: null,
      zip: null
    },
    loading: true,
    charities: []
  },
  created() {
    console.log("created")
    try {
      donate = JSON.parse(localStorage.getItem('donate'));
      if (donate != null) {
        this.donate = donate
        fetch('/api/v1/donate/search', {
          method: 'post',
          body: JSON.stringify(this.donate)
        }).then(response => {
          response.json().then(charities => {
            this.charities = charities
            this.loading = false
            console.log(charities)
          })
        })
      }
    } catch(e) {
      localStorage.removeItem('donate')
    }
  },
  methods:{
  }
});
