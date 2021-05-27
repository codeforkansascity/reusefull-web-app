Vue.component('charity-card', {
  props: { charity: Object },
  computed: {
    charityLink() {
      return `/charity/${this.charity.id}`;
    },
    phoneLink() {
      return `tel:${this.charity.phone}`;
    },
    phone() {
      const clean = this.charity.phone.replace(/[^\d]/g, '');
      return (
        `(${clean.slice(0, 3)})` +
        ` ${clean.slice(3, 6)}` +
        `-${clean.slice(6, 10)}`
      );
    },
    mapLink() {
      return (
        `https://www.google.com/` +
        `maps/dir/?api=1&destination=` +
        `${encodeURIComponent(this.charity.address)}`
      );
    },
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
          <em>{{charity.address}}, {{charity.city}}, {{charity.state}} {{charity.zip}}</em>
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
    `,
});

const app = new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
    errors: [],
    donate: {
      resell: false,
      newItems: false,
      faith: false,
      itemTypes: [],
      charityTypes: [],
      anyCharityType: null,
      budget: null,
      pickupDropoff: null,
      zip: null,
    },
    loading: true,
    charities: [],
  },
  created() {
    try {
      donate = JSON.parse(localStorage.getItem('donate'));
      if (donate != null) {
        this.donate = donate;
        fetch('/api/v1/donate/search', {
          method: 'post',
          body: JSON.stringify(this.donate),
        }).then((response) => {
          response.json().then((charities) => {
            this.charities = charities;
            this.loading = false;

            console.log("creating new map")
            var map = new mapboxgl.Map({
              container: 'map', // container id
              accessToken: 'pk.eyJ1IjoiaHlwcm5pY2siLCJhIjoiY2ttYTBidnYyMW45dTJ2cGJxbmxjMGsyMiJ9.po3lOo4mj9GAEdBBnMjDLA',
              style: 'mapbox://styles/mapbox/streets-v11', // style URL
              center: [-94.57, 39.12],
              zoom: 9 // starting zoom
            });

            for (const c of charities) {
              console.log(c)
              var el = document.createElement('div');
              el.className = 'marker'

              // make a marker for each feature and add to the map
              new mapboxgl.Marker(el)
                .setLngLat([c.lng, c.lat])
                .setPopup(new mapboxgl.Popup({ offset: 25 }) // add popups
                  .setHTML(`
                    <div class="card">
                      <div class="text-center">
                        <img src="${c.logoURL}" class="card-img-top" style="width: 100px;">
                      </div>
                      <div class="card-body">
                        <h5 class="card-title"><a href="/charity/${c.id}">${c.name}</a></h5>
                        <p class="card-text">${c.mission}</p>
                        <p class="card-text"><small class="text-muted">3 miles away</small></p>
                      </div>
                    </div>
                  `)
                )
                .addTo(map)
            }


          });
        });
      }
    } catch (e) {
      localStorage.removeItem('donate');
    }
  },
  methods: {},
});
