Vue.component("charity-card", {
    props: { charity: Object },
    computed: {
        charityLink() {
            return `/charity/${this.charity.id}`;
        },
        phoneLink() {
            return `tel:${this.charity.phone}`;
        },
        phone() {
            const clean = this.charity.phone.replace(/[^\d]/g, "");
            return (
                `(${clean.slice(0, 3)})` +
                ` ${clean.slice(3, 6)}` +
                `-${clean.slice(6, 10)}`
            );
        },
        charityWebsite() {
            if (this.charity.website.slice(0, 4) !== "http") {
                return "http://" + this.charity.website;
            }
            return this.charity.website;
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
    <div class="col-12 col-lg-6 py-3 mb-3">
        <div class="card" style="width: 100%">
            <div class="d-flex flex-column justify-content-center px-3 pt-5 pb-3">
                <a :href="charityLink" class="d-flex card-logo align-self-center justify-content-center mb-sm-2 mr-md-2">
                    <img
                        :src="charity.logoURL"
                        class="img-fluid w-75"
                        alt="Org logo"
                    />
                </a>
                <div class="card-body">
                    <h5 class="card-title mb-3"><a :href="charityLink" style="text-decoration: none;">{{ charity.name }}</a></h5>
                    <div class="my-1"><strong>Pick-Up:</strong>
                        <i class="bi bi-check-circle" style="color: green" v-if="charity.pickup"></i>
                        <i class="bi bi-x-circle" style="color: red" v-if="!charity.pickup"></i>
                    <strong style="margin-left: 1rem;">Dropoff:</strong>
                        <i class="bi bi-check-circle" style="color: green" v-if="charity.dropoff"></i>
                        <i class="bi bi-x-circle" style="color: red" v-if="!charity.dropoff"></i>
                    </div>
                    <div>
                        <i class="bi bi-geo-alt-fill contrast-icon"></i>
                        <strong>Address:</strong>
                        <em>{{charity.address}}, {{charity.city}}, {{charity.state}} {{charity.zip}}</em>
                        <a target="_blank" :href="mapLink">Directions</a>
                    </div>
                    <div class="w-100 d-flex flex-column flex-lg-row">
                        <div class="me-3">
                            <i class="bi bi-telephone-fill contrast-icon"></i>
                            <strong>Phone:</strong>
                            <a class="charityTelephoneLink" :href="phoneLink">{{ phone }}</a>
                        </div>
                        <div>
                            <i class="bi bi-globe contrast-icon"></i>
                            <strong>Website:</strong>
                            <a :href="charityWebsite">{{ charityWebsite }}</a>
                        </div>
                    </div>
                    <p class="mr-auto mt-4 mb-0 pb-0 org-card-mission" v-if="charity.mission">
                        {{ charity.mission }}
                    </p>
                </div>
            </div>
        </div>
    </div>
    `,
});

const app = new Vue({
    delimiters: ["${", "}"],
    el: "#app",
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
            pickup: null,
            dropoff: null,
            zip: null,
        },
        loading: true,
        charities: [],
    },
    created() {
        try {
            donate = JSON.parse(localStorage.getItem("donate"));
            if (donate != null) {
                this.donate = donate;

                // emit an event for analytics to track how often an item category is searched for
                window.dataLayer = window.dataLayer || [];
                for (item of this.donate.itemTypes) {
                    window.dataLayer.push({
                        event: "item_category_search",
                        itemCategory: item,
                    });
                }
                fetch("/api/v1/donate/search", {
                    method: "post",
                    body: JSON.stringify(this.donate),
                }).then((response) => {
                    response.json().then((charities) => {
                        this.charities = charities;
                        this.loading = false;

                        console.log("creating new map");
                        var map = new mapboxgl.Map({
                            container: "map", // container id
                            accessToken:
                                "pk.eyJ1IjoiaHlwcm5pY2siLCJhIjoiY2ttYTBidnYyMW45dTJ2cGJxbmxjMGsyMiJ9.po3lOo4mj9GAEdBBnMjDLA",
                            style: "mapbox://styles/mapbox/streets-v11", // style URL
                            center: [-94.57, 39.12],
                            zoom: 9, // starting zoom
                        });

                        // Shuffle the charities array
                        charities = charities.sort(() => Math.random() - 0.5);

                        for (const c of charities) {
                            var el = document.createElement("div");
                            el.className = "marker";

                            // make a marker for each feature and add to the map
                            new mapboxgl.Marker(el)
                                .setLngLat([c.lng, c.lat])
                                .setPopup(
                                    new mapboxgl.Popup({ offset: 25 }) // add popups
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
                                .addTo(map);
                        }
                    });
                });
            }
        } catch (e) {
            localStorage.removeItem("donate");
        }
    },
    methods: {},
});
